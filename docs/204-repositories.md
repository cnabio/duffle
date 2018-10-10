# Duffle Repositories

A Duffle repository is a location where packaged bundles can be stored and shared. They can be
created by anyone to distribute their own bundles, and users can extend Duffle's list of available
bundles by adding the repository using the `duffle repo` tool.

This section explains how to create and work with Duffle repositories.

## Repository Specification

This specification refers to the HTTP server architecture, but you can use any protocol as long as
it’s a protocol that Duffle understands (which - at the time of writing this doc - is HTTP).

### Hosting Repositories

Repositories must be able to handle the following endpoints:

- `GET /repositories/namespace/name/tags/tag HTTP/1.1`
- `POST /repositories/namespace/name/tags/tag HTTP/1.1`
- `GET /index.yaml HTTP/1.1`

Where

- namespace is an alphanumeric organization name (OPTIONAL)
- name is a name for a given bundle
- tag is a tagged version of a bundle
- HTTP GET requests are used for retrieving bundles
- HTTP POST requests are used for uploading bundles
- `GET /index.yaml` is used to fetch the index (see [Searching](#searching))

For example, with the following command:

```bash
duffle install foo example.com/username/foo:latest
```

Duffle will try to fetch metadata of the bundle using

```text
GET /repositories/username/foo/tags/latest HTTP/1.1
```

Similarly, if the user requests a different tag, like

```bash
duffle install foo example.com/username/foo:1.0.0
```

Duffle will try to fetch metadata of the bundle using

```text
GET /repositories/username/foo/tags/1.0.0 HTTP/1.1
```

NOTE: If tag is not present with `duffle install`, the "latest" tag is implied.

To configure an ordinary web server to serve these bundles, you merely need to do the following:

- Put your bundles in a directory that the server can serve
- Make sure json files are served with the correct content type (application/json)

For example, if you want to serve your bundles out of `$WEBROOT/`, put the bundles and their metadata inside of that folder.

### Bundle Metadata

Duffle fetches the bundles straight from the repository. When you fetch the bundle metadata via

```text
GET /repositories/username/foo/tags/latest HTTP/1.1
```

See [The bundle.json File](101-bundle-json.md) for more information on the accepted types of bundles stored at this endpoint.

## Role-Based Access Control

TODO

## Searching

Searching across repositories is implemented through the `duffle search` command. At runtime, `duffle search` will fetch the index of any logged in repositories using

```text
GET /index.json HTTP/1.1
```

Using this file, `duffle search` can get an idea on what bundles are available in the repository and report back to the user based on the requested keywords.

The index file is a JSON file. It contains metadata about the bundles. The index file contains information about each bundle in the repository. A valid bundle repository must have an index file.

Here is an example of an index file:

```json
{
    "apiVersion": "v1",
    "entries": {
        "aks-bundle": [
            {
                "credentials": {
                    "subscription": {
                        "env": "SUBSCRIPTION",
                        "path": ""
                    }
                },
                "description": "",
                "images": null,
                "invocationImage": {
                    "image": "michellenoorali/aks-bundle:0.2.0",
                    "imageType": "docker"
                },
                "keywords": null,
                "maintainers": null,
                "name": "aks-bundle",
                "parameters": {
                    "domain": {
                        "defaultValue": "containernativelabs.io",
                        "metadata": {},
                        "type": "string"
                    },
                    "lego_email": {
                        "defaultValue": "minooral@microsoft.com",
                        "metadata": {},
                        "type": "string"
                    },
                    "resource_group": {
                        "defaultValue": "duffle-aks",
                        "metadata": {},
                        "type": "string"
                    }
                },
                "version": "0.2.0"
            }
        ]
    },
    "generated": "2018-09-25T12:37:06.336758641-07:00"
}
```

It is the concern of the repository maintainer to generate this index file based on the contents of
its repository.

## Motivation

Duffle repositories are a centralized location where packaged bundles can be stored and shared. They can be created by anyone to distribute their own bundles, and users can bring in these repositories to extend Duffle's list of installable bundles. It makes searching, fetching and hosting bundles easier for both the producer and the consumer of these bundles.

## Rationale

In early versions of Duffle, we experimented with repositories being hosted from git repositories. We swapped this out with the HTTP-based approach after feedback from Independent Software Vendors (ISVs) for a couple reasons:

- Scalability: ISVs can leverage their existing "object storage" platforms for hosting bundles (e.g. Azure Artifacts, Google Cloud Storage, AWS S3)
- Ecosystem compliance: Duffle bundle repositories align more closely with similar distribution models like [Docker's distribution project](https://github.com/docker/distribution)

## Reference Implementations

Duffle should be able to handle the entire workflow for preparing a set of bundles to be hosted on a server, but leaves the hosting up to the individual. It should be able to handle the entire lifecycle from logging into a repository, setting up the development environment, generating a bundle for publishing, and fetching the bundles from the repository.

For the PoC, Duffle will implement the following commands to handle the bundle repository's lifecycle:

- `duffle login` logs in to a bundle repository
- `duffle logout` logs out from a bundle repository
- `duffle pull` pulls bundles from a bundle repository
- `duffle push` pushes bundles to a bundle repository
- `duffle search` searches across logged in bundle repositories for bundles
