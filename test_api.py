import requests
import json

BASE_URL = "http://localhost:8350"

def test_health():
    """Test health check endpoint"""
    print("=" * 50)
    print("Testing Health Check...")
    print("=" * 50)
    response = requests.get(f"{BASE_URL}/healthcheck")
    print(f"Status Code: {response.status_code}")
    print(f"Response: {response.text}")
    print()

def test_authenticate():
    """Test device authentication"""
    print("=" * 50)
    print("Testing Device Authentication...")
    print("=" * 50)
    
    url = f"{BASE_URL}/v2/rpc/authenticate_device?unwrap"
    payload = {"device_id": "test-player-001"}
    headers = {"Content-Type": "application/json"}
    
    response = requests.post(url, json=payload, headers=headers)
    print(f"Status Code: {response.status_code}")
    
    if response.status_code == 200:
        data = response.json()
        print(f"✅ Authentication Successful!")
        print(f"User ID: {data.get('user_id')}")
        print(f"Username: {data.get('username')}")
        print(f"Session Token: {data.get('session_token')[:50]}..." if data.get('session_token') else "No token")
        
        # Print profile
        profile = data.get('profile', {})
        print(f"\nPlayer Profile:")
        print(f"  Wins: {profile.get('wins', 0)}")
        print(f"  Losses: {profile.get('losses', 0)}")
        print(f"  Draws: {profile.get('draws', 0)}")
        print(f"  Rating: {profile.get('rating', 1000)}")
        print()
        return data.get('session_token')
    else:
        print(f"❌ Authentication Failed")
        print(f"Response: {response.text}")
        print()
        return None

def test_leaderboard(token):
    """Test get leaderboard endpoint"""
    if not token:
        print("⚠️  Skipping leaderboard test - no auth token")
        return
        
    print("=" * 50)
    print("Testing Get Leaderboard...")
    print("=" * 50)
    
    url = f"{BASE_URL}/v2/rpc/get_leaderboard?unwrap"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }
    
    response = requests.get(url, headers=headers)
    print(f"Status Code: {response.status_code}")
    
    if response.status_code == 200:
        data = response.json()
        entries = data.get('entries', [])
        print(f"✅ Leaderboard Retrieved!")
        print(f"Total Entries: {len(entries)}")
        
        if entries:
            print("\nTop Players:")
            for i, entry in enumerate(entries[:5], 1):
                print(f"  {i}. {entry.get('username')} - Rating: {entry.get('score')}")
        else:
            print("  No entries yet")
        print()
    else:
        print(f"❌ Failed to get leaderboard")
        print(f"Response: {response.text}")
        print()

def main():
    print("\n🎮 Tic-Tac-Toe Nakama Backend API Tests\n")
    
    # Test 1: Health Check
    test_health()
    
    # Test 2: Authentication
    token = test_authenticate()
    
    # Test 3: Leaderboard
    test_leaderboard(token)
    
    print("=" * 50)
    print("✅ All tests completed!")
    print("=" * 50)

if __name__ == "__main__":
    main()
