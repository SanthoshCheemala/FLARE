const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "/api";

// Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: {
    id: number;
    email: string;
    role: string;
  };
}

export interface StartScreeningRequest {
  name: string;
  customerListId: number;
  sanctionListIds: number[];
  columnMapping?: Record<string, string>;
}

export interface StartScreeningResponse {
  jobId: string;
}

export interface ScreeningProgress {
  phase: string;
  stage:
    | "preparing"
    | "init_server"
    | "encrypting"
    | "intersecting"
    | "complete"
    | "failed";
  message: string;
  percent: number;
  metrics?: Record<string, string>;
}

export interface ScreeningStatus {
  jobId: string;
  status: string;
  stage: string;
}

export interface CustomerList {
  id: number;
  name: string;
  description: string;
  recordCount: number;
  createdAt: string;
}

export interface SanctionList {
  id: number;
  name: string;
  source: string;
  description: string;
  recordCount: number;
  createdAt: string;
}

export interface DashboardStats {
  totalScreenings: number;
  totalMatches: number;
  activeLists: number;
  recentScreenings: Array<{
    id: number;
    jobId: string;
    name: string;
    status: string;
    matchCount: number;
    finishedAt: string;
    createdAt: string;
  }>;
  systemStatus: string;
  activeWorkers: number;
}

// Token management
let accessToken: string | null = null;

export function setAccessToken(token: string) {
  accessToken = token;
  if (typeof window !== "undefined") {
    localStorage.setItem("accessToken", token);
  }
}

export function getAccessToken(): string | null {
  if (!accessToken && typeof window !== "undefined") {
    accessToken = localStorage.getItem("accessToken");
  }
  return accessToken;
}

export function clearAccessToken() {
  accessToken = null;
  if (typeof window !== "undefined") {
    localStorage.removeItem("accessToken");
    localStorage.removeItem("user");
  }
}

export function getUser(): { name: string; email: string; role: string } | null {
  if (typeof window !== "undefined") {
    const userStr = localStorage.getItem("user");
    if (userStr) {
      return JSON.parse(userStr);
    }
  }
  return null;
}

// API Client
class APIClient {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  private async handleResponse<T>(
    response: Response,
    isLoginRequest: boolean = false
  ): Promise<T> {
    if (!response.ok) {
      if (response.status === 401 && !isLoginRequest) {
        clearAccessToken();
        if (
          typeof window !== "undefined" &&
          !window.location.pathname.includes("/login")
        ) {
          window.location.href = "/login";
        }
      }
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }
    return response.json();
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const token = getAccessToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    if (options.headers) {
      Object.assign(headers, options.headers);
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    });

    return this.handleResponse<T>(response, endpoint.includes("/auth/login"));
  }

  // Auth
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    });
    setAccessToken(response.accessToken);
    if (typeof window !== "undefined") {
      localStorage.setItem("user", JSON.stringify({
        name: "Admin User", // Backend doesn't send name yet, defaulting
        email: response.user.email,
        role: response.user.role
      }));
    }
    return response;
  }

  // Screenings
  async startScreening(
    request: StartScreeningRequest
  ): Promise<StartScreeningResponse> {
    return this.request<StartScreeningResponse>("/screenings", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  // Lists
  async getCustomerLists(): Promise<CustomerList[]> {
    return this.request<CustomerList[]>("/lists/customers");
  }

  async deleteCustomerList(id: number): Promise<void> {
    return this.request(`/lists/customers/${id}`, {
      method: "DELETE",
    });
  }

  async getCustomerListHeaders(id: string): Promise<string[]> {
    const data = await this.request<{ headers: string[] }>(`/lists/customers/${id}/headers`);
    return data.headers;
  }

  async getSanctionLists(): Promise<SanctionList[]> {
    return this.request<SanctionList[]>("/lists/sanctions");
  }

  async deleteSanctionList(id: number): Promise<void> {
    return this.request(`/lists/sanctions/${id}`, {
      method: "DELETE",
    });
  }

  async uploadCustomerList(file: File, name: string, description: string) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("name", name);
    formData.append("description", description);

    const token = getAccessToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${this.baseURL}/lists/customers/upload`, {
      method: "POST",
      headers,
      body: formData,
    });

    return this.handleResponse(response);
  }

  async uploadSanctionList(
    file: File,
    name: string,
    source: string,
    description: string
  ) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("name", name);
    formData.append("source", source);
    formData.append("description", description);

    const token = getAccessToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${this.baseURL}/lists/sanctions/upload`, {
      method: "POST",
      headers,
      body: formData,
    });

    return this.handleResponse(response);
  }

  async getScreeningStatus(jobId: string): Promise<ScreeningStatus> {
    return this.request<ScreeningStatus>(`/screenings/${jobId}/status`);
  }

  // SSE for real-time progress
  subscribeToScreeningEvents(
    jobId: string,
    onProgress: (progress: ScreeningProgress) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ): () => void {
    const token = getAccessToken();
    const url = `${this.baseURL}/screenings/${jobId}/events`;
    let eventSource: EventSource | null = null;
    let jobCompleted = false;
    let retryCount = 0;
    const maxRetries = 5;
    let retryTimeout: NodeJS.Timeout | null = null;

    const connect = () => {
      if (jobCompleted) return;

      eventSource = new EventSource(token ? `${url}?token=${token}` : url);

      eventSource.addEventListener("done", (event) => {
        jobCompleted = true;
        eventSource?.close();
        onComplete?.();
      });

      eventSource.onmessage = (event) => {
        // Reset retry count on successful message
        retryCount = 0;

        if (event.data === ": connected") {
          return;
        }

        try {
          const progress: ScreeningProgress = JSON.parse(event.data);
          onProgress(progress);

          if (progress.stage === "complete" || progress.stage === "failed") {
            jobCompleted = true;
            eventSource?.close();
          }
        } catch (error) {
          console.error("Failed to parse SSE event:", error);
        }
      };

      eventSource.onerror = (error) => {
        if (jobCompleted) {
          eventSource?.close();
          return;
        }

        // Connection lost
        eventSource?.close();
        
        if (retryCount < maxRetries) {
          retryCount++;
          const delay = Math.min(1000 * Math.pow(2, retryCount), 10000); // Exponential backoff
          retryTimeout = setTimeout(connect, delay);
        } else {
          console.error("SSE connection failed after max retries");
          onError?.(new Error("Connection lost"));
        }
      };
    };

    connect();

    // Return cleanup function
    return () => {
      jobCompleted = true;
      if (retryTimeout) clearTimeout(retryTimeout);
      if (eventSource) eventSource.close();
    };
  }

  async getScreeningResults(
    jobId: string,
    limit: number = 50,
    offset: number = 0
  ) {
    return this.request(
      `/screenings/${jobId}/results?limit=${limit}&offset=${offset}`
    );
  }

  async updateResultStatus(
    resultId: number,
    status: "PENDING" | "CONFIRMED" | "FALSE_POSITIVE"
  ): Promise<void> {
    return this.request(`/results/${resultId}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  }

  // Health check
  async healthCheck(): Promise<{ status: string }> {
    const response = await fetch(`${this.baseURL}/health`);
    if (!response.ok) {
      throw new Error("Health check failed");
    }
    return { status: "OK" };
  }

  async getDashboardStats(): Promise<DashboardStats> {
    return this.request<DashboardStats>("/dashboard/stats");
  }

  async getPerformanceMetrics(): Promise<{
    performance: {
      total_time_seconds: number;
      total_time_formatted: string;
      key_gen_time_seconds: number;
      key_gen_time_formatted: string;
      key_gen_percent: number;
      hashing_time_seconds: number;
      hashing_time_formatted: string;
      hashing_percent: number;
      witness_time_seconds: number;
      witness_time_formatted: string;
      witness_percent: number;
      intersection_time_seconds: number;
      intersection_time_formatted: string;
      intersection_percent: number;
      num_workers: number;
      total_operations: number;
      throughput_ops_per_sec: number;
    };
    memory: {
      alloc_mb: number;
      total_alloc_mb: number;
      sys_mb: number;
      num_gc: number;
      goroutines: number;
    };
  }> {
    return this.request("/performance/metrics");
  }
}

// Export singleton instance
export const apiClient = new APIClient(API_BASE_URL);
