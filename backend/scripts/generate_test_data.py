#!/usr/bin/env python3
"""
Generate larger test datasets to reduce PSI collision probability.
Target: ~1000 records to bring collision probability below 5%
"""

import csv
import random
from datetime import datetime, timedelta

# Sample data pools
FIRST_NAMES = [
    "James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
    "William", "Barbara", "David", "Elizabeth", "Richard", "Susan", "Joseph", "Jessica",
    "Thomas", "Sarah", "Christopher", "Karen", "Daniel", "Nancy", "Matthew", "Lisa",
    "Anthony", "Betty", "Mark", "Margaret", "Donald", "Sandra", "Steven", "Ashley",
    "Paul", "Kimberly", "Andrew", "Emily", "Joshua", "Donna", "Kenneth", "Michelle",
    "Christina", "George", "Laura", "Kevin", "Carol", "Brian", "Amanda", "Edward"
]

LAST_NAMES = [
    "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
    "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
    "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
    "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
    "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
    "Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
    "Carter", "Roberts", "Vargas", "Harper", "Newman", "Blake", "Walters"
]

COUNTRIES = [
    "US", "UK", "CA", "AU", "DE", "FR", "IT", "ES", "NL", "SE",
    "NO", "DK", "FI", "IE", "NZ", "CH", "AT", "BE", "PT", "GR",
    "PL", "CZ", "HU", "RO", "BG", "HR", "SI", "SK", "LT", "LV",
    "EE", "JP", "KR", "CN", "IN", "SG", "MY", "TH", "PH", "ID",
    "VN", "BR", "MX", "AR", "CL", "CO", "PE", "VE", "ZA", "EG"
]

PROGRAMS = [
    "OFAC SDN", "EU Sanctions", "UN Sanctions", "DFAT Sanctions",
    "HMT Sanctions", "SECO Sanctions", "FINMA Watchlist", "PEP List"
]

def random_date(start_year=1940, end_year=2005):
    """Generate random date between start_year and end_year"""
    start = datetime(start_year, 1, 1)
    end = datetime(end_year, 12, 31)
    delta = end - start
    random_days = random.randint(0, delta.days)
    date = start + timedelta(days=random_days)
    return date.strftime("%Y-%m-%d")

def generate_customers(count=1000, include_christina=True):
    """Generate customer records"""
    customers = []
    
    # Always include Christina Vargas as first record if requested
    if include_christina:
        customers.append({
            "name": "Christina Vargas",
            "date_of_birth": "1995-07-23",
            "country": "AU"
        })
    
    # Generate remaining records
    for i in range(count - (1 if include_christina else 0)):
        name = f"{random.choice(FIRST_NAMES)} {random.choice(LAST_NAMES)}"
        dob = random_date()
        country = random.choice(COUNTRIES)
        
        customers.append({
            "name": name,
            "date_of_birth": dob,
            "country": country
        })
    
    return customers

def generate_sanctions(count=1000, include_christina=True):
    """Generate sanction records"""
    sanctions = []
    
    # Always include Christina Vargas as first record if requested
    if include_christina:
        sanctions.append({
            "name": "Christina Vargas",
            "date_of_birth": "1995-07-23",
            "country": "AU",
            "program": "OFAC SDN",
            "aliases": ""
        })
    
    # Generate remaining records
    for i in range(count - (1 if include_christina else 0)):
        name = f"{random.choice(FIRST_NAMES)} {random.choice(LAST_NAMES)}"
        dob = random_date()
        country = random.choice(COUNTRIES)
        program = random.choice(PROGRAMS)
        
        # 10% chance of having an alias
        aliases = ""
        if random.random() < 0.1:
            alias_name = f"{random.choice(FIRST_NAMES)} {random.choice(LAST_NAMES)}"
            aliases = alias_name
        
        sanctions.append({
            "name": name,
            "date_of_birth": dob,
            "country": country,
            "program": program,
            "aliases": aliases
        })
    
    return sanctions

def write_csv(filename, data, fieldnames):
    """Write data to CSV file"""
    with open(filename, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(data)
    print(f"✅ Generated {filename} with {len(data)} records")

if __name__ == "__main__":
    # Generate customer data (1000 records with 1 match)
    customers = generate_customers(1000, include_christina=True)
    write_csv("client_data_large.csv", customers, ["name", "date_of_birth", "country"])
    
    # Generate sanction data (1000 records with 1 match)
    sanctions = generate_sanctions(1000, include_christina=True)
    write_csv("server_data_large.csv", sanctions, ["name", "date_of_birth", "country", "program", "aliases"])
    
    print("\n📊 Summary:")
    print(f"  - Customers: {len(customers)} (including Christina Vargas)")
    print(f"  - Sanctions: {len(sanctions)} (including Christina Vargas)")
    print(f"  - Expected matches: 1 (Christina Vargas)")
    print(f"\n💡 Upload these files through the UI to test with lower collision probability")
