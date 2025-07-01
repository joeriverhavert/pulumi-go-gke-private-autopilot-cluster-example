# Pulumi GCP Go: Example GKE Pilot Cluster  

This template provisions a **Google Kubernetes Engine (GKE) Autopilot Private Cluster** using Pulumi and Go. It also stores the generated **kubeconfig as a Secret** in Google Cloud Secret Manager.

It demonstrates how to:

- Use the Pulumi GCP provider in a Go program
- Provision a secure GKE Autopilot cluster with private nodes and control plane
- Generate and store the cluster kubeconfig as a managed secret

It’s a great starting point for learning Pulumi with Go on GCP or bootstrapping a private, production-grade Kubernetes environment.

---

## 📦 Providers

- **Google Cloud Platform** via the Pulumi GCP SDK for Go  
  [`github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp`](https://github.com/pulumi/pulumi-gcp)

---

## ☁️ Resources

- **GKE Autopilot Cluster** (`gcp.container.Cluster`)
  - Autopilot mode enabled
  - Private nodes and endpoint access
- **Service Account** (`gcp.serviceaccount.Account`)
  - Used for GKE node pool management
- **Kubeconfig Secret** (`gcp.secretmanager.Secret` + `SecretVersion`)
  - Contains `kubectl`-ready config stored securely

---

## 🔁 Outputs

- **Cluster Name**: The GKE cluster's name
- **Cluster Endpoint**: HTTPS address for API access
- **Kubeconfig Secret Name**: The Secret Manager name storing the kubeconfig
- **Kubeconfig Secret Version Name**: The Secret Version name of storing the kubeconfig 

---

## 📌 When to Use This Template

Use this if:

- You want a **minimal, secure GKE Autopilot cluster** example
- You need to **store kubeconfig safely** and retrieve it programmatically
- You're exploring Pulumi’s Go SDK with GCP infrastructure
- You're building CI/CD or IaC automation with secure access to clusters

---

## 🧰 Prerequisites

- Go 1.20 or later installed
- A Google Cloud project with billing enabled
- GCP credentials configured for Pulumi:

```bash
gcloud auth application-default login
```

## 🚀 Usage
1. Scaffold your project
```bash
pulumi new gcp-go
```

2. Set your project ID:
```bash
pulumi config set gcp:project conro-sbx
```

3. Set the Google Cloud Platform region(optional)
```bash
pulumi config set gcp:region europe-west1
```

4. Preview and deploy the resources
```bash
pulumi preview
pulumi up
```

## 🗂️ Project Layout
```bash
├── Pulumi.yaml             # Project metadata
├── Pulumi.<stack>.yaml     # Stack-specific configuration
├── go.mod                  # Go module dependencies
└── main.go                 # Pulumi program: provisions GKE and stores kubeconfig
```

 ## 📚 Getting Help

 - Pulumi Documentation: https://www.pulumi.com/docs/
 - GCP Provider Reference: https://www.pulumi.com/registry/packages/gcp/
 - Community Slack: https://slack.pulumi.com/
 - GitHub Issues: https://github.com/pulumi/pulumi/issues adjust my readme.
