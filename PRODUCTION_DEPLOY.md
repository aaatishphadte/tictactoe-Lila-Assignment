# Production Deployment Guide (Google Cloud)

Follow these steps to deploy your Tic-Tac-Toe game to Google Cloud Platform.

## 1. Backend: Google Compute Engine (VM)

### Create the VM
1. Go to the [GCP Console](https://console.cloud.google.com/).
2. Create a Compute Engine instance:
   - **Machine type**: `e2-medium` (recommended) or `e2-small`.
   - **Boot disk**: Ubuntu 22.04 LTS.
   - **Firewall**: Allow HTTP and HTTPS traffic.

### Configure Firewall
1. Open the [Firewall Rules](https://console.cloud.google.com/net-security/firewall/networks) page.
2. Create a rule called `nakama-ports`:
   - **Targets**: All instances in the network.
   - **Source IP ranges**: `0.0.0.0/0`.
   - **Protocols and ports**: TCP `7349, 7350, 7351, 8350`.

### Setup VM
Once you SSH into your VM, run:
```bash
sudo apt-get update
sudo apt-get install -y docker.io docker-compose
```

### Deploy
1. Upload the `nakama` and `modules` folders to the VM.
2. Run:
   ```bash
   cd nakama
   sudo docker-compose up -d --build
   ```

## 2. Frontend: Firebase Hosting

### Build locally
1. Go to the `client` folder.
2. Rename `.env.production.example` to `.env.production`.
3. Update `VITE_NAKAMA_HOST` with your VM's **External IP**.
4. Run:
   ```bash
   npm install
   npm run build
   ```

### Deploy
1. Install Firebase CLI: `npm install -g firebase-tools`.
2. Login: `firebase login`.
3. Initialize: `firebase init hosting` (use the existing files I created).
4. Deploy: `firebase deploy`.

---
Your game will now be live at your Firebase project URL!
