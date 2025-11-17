# **ComposePack**

> 🧩 **Templating for Docker Compose — finally done right.**
> The power of Helm-style configuration for Docker Compose.

<p align="center">
  <!-- Optional: drop a banner here later -->
  <!-- <img src="docs/images/banner.png" width="700" /> -->
</p>

Docker Compose is great for running containers —
but it has one huge missing piece:
**no templating, no dynamic config, no values, no clean overrides.**

This forces teams to ship giant YAML files, copy/paste configs across environments, switch between profiles, hand-edit deployments, extra scripts for variable processing and pray nothing breaks when customers edit their .env files.

**ComposePack fixes all of this.** ✨

ComposePack brings a **modern templating engine**, **values.yaml**, and a **real packaging workflow** to Docker Compose —
all while staying 100% compatible with the Compose CLI.

Think of it as:

<p align="center">
  <b>⚓ Helm-style workflows • 🎛️ Dynamic templating • 📦 Installable charts</b><br>
  <b>→ for Docker Compose ←</b>
</p>

With ComposePack you can:

* 📝 Write Compose files using **Go-style templates**
* ⚙️ Ship clean **values.yaml** defaults + user overrides
* 📦 Distribute your app as an **installable chart**
* 🔐 Render an isolated, reproducible **release directory**
* 🧩 Generate a single merged `docker-compose.yaml` at runtime
* 🚀 Run everything through a simple CLI (`install`, `up`, `down`, `logs`, `ps`)

All powered by the tools you already use — **`docker compose` under the hood**.

```bash
composepack install ./charts/myapp --name prod -f values-prod.yaml --auto-start
```

Whether you're shipping on-prem software, managing multi-env stacks, or just sick of duplicating Compose and .env files, ComposePack brings structure, clarity, and modern tooling to the Compose ecosystem.

## ⚖️ ComposePack vs. Docker Compose

| What you get                                     | Docker Compose | **ComposePack** |
| ------------------------------------------------ | :------------: | :-------------: |
| Templating for Compose files                     |       ❌        |      **✅**      |
| Structured config model (system vs. user values) | ❌ (flat .env)  |      **✅**      |
| Installable packages (charts)                    |       ❌        |      **✅**      |
| Reproducible release environments                |       ❌        |      **✅**      |
| 100% Compose-compatible runtime                  |       ✅        |      **✅**      |
