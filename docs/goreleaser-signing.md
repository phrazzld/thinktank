# GoReleaser Signing Configuration

This document explains how the signing process works in GoReleaser for the thinktank project.

## Overview

GoReleaser is configured to sign release artifacts (specifically checksums) when creating official releases. This signing process:

1. Adds a layer of security by verifying the authenticity of released binaries
2. Uses GPG to create detached signatures for checksums
3. Only applies to official tagged releases, not snapshot or development builds

## Signing Configuration

The signing configuration is defined in `.goreleaser.yml` under the `signs` section:

```yaml
# Signs the checksum file using GPG (for official releases)
# For snapshot builds, use --skip=sign in the command line
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sig"
      - "${artifact}"
```

## How Signing Works

1. For official releases (tagged releases):
   - GPG signing is enabled when the `GPG_FINGERPRINT` environment variable is set
   - The checksums file is signed with a detached signature
   - The signature is published alongside the release artifacts

2. For snapshot builds (PRs, development builds):
   - Signing is explicitly disabled using the `--skip=sign` command line flag
   - This prevents errors when GPG keys aren't available in CI environments

## Required Environment Variables

For signing to work, the following environment variables must be set:

- `GPG_FINGERPRINT`: The fingerprint of the GPG key to use for signing
- `GPG_PASSWORD` (optional): The password for the GPG key, if required

These are typically set as repository secrets for official release builds.

## CI/CD Integration

In the release workflow (`.github/workflows/release.yml`), different flags are used based on the release type:

- For official releases:
  ```bash
  goreleaser release --release-notes=CHANGELOG.md --clean
  ```

- For snapshot builds:
  ```bash
  goreleaser release --snapshot --skip-publish --skip-announce --skip-sign --release-notes=CHANGELOG.md --clean
  ```

Note the `--skip-sign` flag for snapshots, which explicitly disables signing.

The workflow uses a specific version of GoReleaser to ensure compatibility:

```yaml
- name: Install GoReleaser
  uses: goreleaser/goreleaser-action@v5
  with:
    version: v1.20.0
    install-only: true
```

## Troubleshooting

If you encounter issues with signing:

1. For local builds, ensure you have GPG properly set up if you want to test signing
2. For CI builds, remember that signing is only enabled for official tagged releases
3. If testing locally, you can use `--skip=sign` to bypass the signing process:
   ```bash
   goreleaser release --snapshot --skip=announce,publish,sign
   ```

## Adding GPG Keys for Official Releases

When setting up official releases:

1. Create a GPG key specifically for the project
2. Export the public and private key
3. Add the private key and fingerprint as repository secrets
4. Configure the workflow to import the key before running GoReleaser
