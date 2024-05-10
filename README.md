# arpload
Simply & safely upload a large file to Arweave network. An interrupted upload is resumed.

```bash
# Install with Go
go install github.com/intob/arpload@latest

# Run
AR_WALLET=~/path/to/wallet.json arpload -type="image/avif" ~/path/to/file
```
