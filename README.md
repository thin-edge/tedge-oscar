# tedge-oscar

A Go CLI tool to manage thin-edge.io flows, including pulling and pushing flow images as OCI artifacts using the oras library, and managing flow instances.

## Commands

- `tedge-oscar flows images pull` — Pull a flow image from an OCI registry
- `tedge-oscar flows images push` — Push a flow image to an OCI registry
- `tedge-oscar flows images list` — List available flow images
- `tedge-oscar flows instances list` — List deployed flow instances
- `tedge-oscar flows instances deploy` — Deploy a flow instance

## Typical Workflow Example

1. Publish (push) a flow image to a registry

   ```sh
   tedge-oscar flows images push ghcr.io/youruser/your-flow:1.0 \
     --file flow.json --file README.md
   ```

2. Pull a flow image from a registry

   ```sh
   tedge-oscar flows images pull ghcr.io/youruser/your-flow:1.0
   ```

   You can also save it as a tarball using

   ```sh
   tedge-oscar flows images pull ghcr.io/youruser/your-flow:1.0 --tarball
   ```

3. Deploy an instance using the pulled image

   ```sh
   tedge-oscar flows instances deploy myinstance ghcr.io/youruser/your-flow:1.0 \
     --topics te/device/main///m/+
   ```

4. List deployed instances

   ```sh
   tedge-oscar flows instances list
   ```

5. Remove an instance

   ```sh
   tedge-oscar flows instances remove myinstance
   ```

## Development

- Built with [Cobra](https://github.com/spf13/cobra) for CLI structure
- Uses [oras](https://github.com/oras-project/oras-go) for OCI artifact operations

## Getting Started

### Install via golang

You can install the CLI directly from the repository using Go 1.21+:

```sh
go install github.com/thin-edge/tedge-oscar/cmd/tedge-oscar@latest
```

This will place the `tedge-oscar` binary in your `$GOBIN` (usually `$HOME/go/bin`). Make sure this directory is in your `$PATH`.

### Build from Source

1. Install Go 1.21 or newer
2. Run `go mod tidy` to install dependencies
3. Build: `go build`
4. Run: `./tedge-oscar [command]`

### References

* https://www.kenmuse.com/blog/universal-packages-on-github-with-oras/

## License

Apache 2.0
