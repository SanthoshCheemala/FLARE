export const APP_MODE = (process.env.NEXT_PUBLIC_APP_MODE || "CLIENT") as "CLIENT" | "SERVER";

export const isClientMode = () => APP_MODE === "CLIENT";
export const isServerMode = () => APP_MODE === "SERVER";
