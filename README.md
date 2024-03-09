# arpload
Simply & safely upload a large file to Arweave network.
An interrupted upload is resumed.

```bash
# Install with Go
go install github.com/intob/arpload

# We need an Arweave wallet
export AR_WALLET=~/path/to/wallet.json

# Run
arpload -type="image/avif" -title=testimg -desc="some test data" -author=joey ~/path/to/img.avif
```