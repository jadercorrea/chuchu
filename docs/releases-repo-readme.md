# GPTCode CLI Releases

Pre-built binaries for the GPTCode CLI agent.

## Quick Install

### macOS / Linux

```bash
curl -sSL https://raw.githubusercontent.com/gptcode-cloud/cli-releases/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
# Download the latest release
$version = (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/gptcode-cloud/cli-releases/main/LATEST").Content.Trim()
$url = "https://github.com/gptcode-cloud/cli-releases/releases/download/$version/gptcode_$($version.TrimStart('v'))_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile gptcode.zip
Expand-Archive -Path gptcode.zip -DestinationPath .
```

## Manual Installation

1. Go to the [Releases](https://github.com/gptcode-cloud/cli-releases/releases) page
2. Download the appropriate archive for your OS/architecture
3. Extract the binary
4. Move it to a directory in your PATH

## Supported Platforms

| OS | Architecture | Archive |
|----|--------------|---------|
| Linux | x86_64 | `gptcode_*_linux_amd64.tar.gz` |
| Linux | ARM64 | `gptcode_*_linux_arm64.tar.gz` |
| macOS | Intel | `gptcode_*_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `gptcode_*_darwin_arm64.tar.gz` |
| Windows | x86_64 | `gptcode_*_windows_amd64.zip` |

## Verify Installation

```bash
gptcode --version
```

## Getting Started

After installation, run:

```bash
gptcode setup
```

This will guide you through configuring your API keys and preferences.

## License

See the main [GPTCode repository](https://github.com/jadercorrea/gptcode) for license information.
