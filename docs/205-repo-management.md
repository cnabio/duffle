# Local Repository Management

This section describes how the Duffle tool interacts with local repositories of CNAB bundles.

All local CNAB bundles are stored in repositories that can be found in `$DUFFLE_HOME/repositories`, following the `registry/<org>/<repository>/tags/<tag>/bundle.json` structure.

> For exported bundles, a thick bundle (containing all assets in a single file) will be stored side-by-side with the bundle file under the corresponding tag.

> The default registry is `https://hub.cnlabs.io/`, so any push / pull operation that doesn't include a registry will be executed against this registry.


## Using the Duffle CLI with local bundles

The Duffle CLI tool is primarily used with bundles found in `$DUFFLE_HOME/repositories`, and commands such as `install`, `push`, `pull` or `tag` will work with repositories there. However, commands that work with bundles also feature a `-f` flag that that helps the use of a bundle in a different place than `$DUFFLE_HOME`, but should be used in exceptional cases.
