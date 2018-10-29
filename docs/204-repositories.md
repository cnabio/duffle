# CNAB Repositories

A CNAB repository is a location where packaged bundles can be stored and shared. They can be created by anyone to distribute their own bundles.

CNAB supports two HTTP based transfer protocols. A "dumb" protocol which requires only a standard HTTP server on the server end of the connection, and a "smart" protocol with deeper integration with CNAB clients. This document describes both protocols.

As a design feature, smart clients can automatically upgrade "dumb" protocol URLs to smart URLs.  This permits all users to have the same published URL, and the peers automatically select the most efficient transport available to them.

This document explains how to create and work with CNAB repositories.

## General Information

This section describes information that applies to both the "smart" and "dumb" protocols.

### URL Format

URLs for CNAB repositories accessed by HTTP use the standard HTTP URL syntax documented by RFC 1738, so they are of the form:

    https://<host>:<port>/<path>?<searchpart>

Within this documentation, the placeholder `$REPO_URL` will stand for the `https://` repository URL entered by the end-user.

It should also be noted that `<path>` can be many levels deep. `/foo` and `/foo/bar/car/star` are both valid paths.

### Authentication

Standard HTTP authentication is used if authentication is required to access a repository, and MAY be configured and enforced by the HTTP server software. Clients SHOULD support Basic authentication as described by RFC 2617. Servers SHOULD support Basic authentication by relying upon the HTTP server placed in front of the CNAB server software.

Clients and servers MAY support other common forms of HTTP based authentication, such as Digest authentication or OAuth2.

### SSL

It is STRONGLY recommended that clients and servers support SSL, particularly to protect passwords during the authentication process.

### Session State

The CNAB transfer protocol is intended to be completely stateless from the perspective of the HTTP server side. All state MUST be retained and managed by the client. This permits simple round-robin load-balancing on the server side, without needing to worry about state management.

### General Request Processing

Except where noted, all standard HTTP behavior SHOULD be assumed by both client and server.  This includes (but is not necessarily limited to):

If there is no repository at `$REPO_URL`, or the resource pointed to by a location matching `$REPO_URL` does not exist, the server MUST NOT respond with `200 OK` response. A server SHOULD respond with `404 Not Found`, `410 Gone`, or any other suitable HTTP status code which does not imply the resource exists as requested.

If there is a repository at `$REPO_URL`, but access is not currently permitted, the server MUST respond with the `403 Forbidden` HTTP status code.

Servers SHOULD support both HTTP 1.0 and HTTP 1.1. Servers SHOULD support chunked encoding for both request and response bodies.

Clients SHOULD support both HTTP 1.0 and HTTP 1.1. Clients SHOULD support chunked encoding for both request and response bodies.

Servers MAY return ETag and/or Last-Modified headers.

Clients MAY revalidate cached entities by including If-Modified-Since and/or If-None-Match request headers.

Servers MAY return `304 Not Modified` if the relevant headers appear in the request and the entity has not changed.  Clients MUST treat `304 Not Modified` identical to `200 OK` by reusing the cached entity.

Clients MAY reuse a cached entity without revalidation if the Cache-Control and/or Expires header permits caching.  Clients and servers MUST follow RFC 2616 for cache controls.

### Server Capabilities

HTTP clients can determine a server's capabilities by making a `GET` request to the root URL, `/`, without any search/query parameters.

The server response SHOULD contain a header called `CNAB-Capabilities` with a whitespace-delimited list of server capabilities. These allow the server to declare what it can and cannot support to the client.

Clients SHOULD fall back to the dumb protocol if the header is not present. When falling back to the dumb protocol, clients SHOULD discard the response already in hand, even if the response code is not between 200-399. Clients MUST NOT continue if they do not support the dumb protocol. Dumb servers MUST NOT return a `CNAB-Capabilities` header.

Servers MUST support all capabilities defined here.

Example smart server reply:

```bash
200 OK
CNAB-Capabilities: "search thin-bundle thick-bundle auth-oauth2"
```

#### search

The "search" capability came about as a way for clients to determine if the repository has a search API available at `/search`, while maintaining compatibility for simpler servers (e.g. a basic file server).

When enabled, this capability means that the server has a search API available.

HTTP clients that support the "smart" protocol (or both the "smart" and "dumb" protocols) MAY search for bundles by making a request to `/search`.

#### thin-bundle

A thin bundle is one which reference container images not contained within the bundle (but are known to exist at the receiving end).

When enabled, this capability means that the server can receive and host thin bundles. Supporting this feature compared to "thick-bundle" can reduce the network traffic significantly.

HTTP servers that support the "smart" protocol MUST support either "thin-bundle" or "thick-bundle", or both.

#### thick-bundle

A thick bundle is one which reference container images are contained within the bundle.

When enabled, this capability means that the server can receive and host thick bundles. Supporting this feature increases the network traffic significantly.

HTTP servers that support the "smart" protocol MUST support either "thin-bundle" or "thick-bundle", or both.

#### auth-oauth2, auth-basic, auth-digest

The "auth-" capabilities came about as a way for clients to determine the authentication strategy to be used against the server for authentication/authorization.

If no "auth-" capability is present, the server supports no auth strategy.

## The "Dumb" Protocol

HTTP clients that only support the "dumb" protocol MUST discover bundles by making a direct request for the bundle of the repository.

Dumb HTTP clients MUST make a `GET` request to `$REPO_URL`, without any search/query parameters.

The Content-Type of the returned entity SHOULD be `multipart/signed; value="application/json"`, but MAY be any content type. Clients MUST attempt to validate the content against the returned Content-Type.

Cache-Control headers MAY be returned to disable caching of the returned entity.

When examining the response, clients SHOULD only examine the HTTP status code. The only valid response codes are `200 OK` and the redirection status codes 300-399; anything else should be considered an error.

The returned content is a clear-signed bundle as described in [The bundle.json File][bundle.json].

If the Content-Type of the returned entity is `multipart/signed; value="application/json"`, clients SHOULD to attempt to verify the bundle's signature. When the signature fails, clients SHOULD NOT continue unless they intentionally choose to ignore the signature failure.

Example dumb server reply:

```bash
200 OK
Content-Type: multipart/signed; value="application/json"

-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

{
  "schemaVersion": "v1",
  "name": "helloworld",
  "version": "0.1.2",
  "description": "An example 'thin' helloworld Cloud-Native Application Bundle",
  ...
}
-----BEGIN PGP SIGNATURE-----
Comment: helloworld v0.1.2

wsBcBAEBCAAQBQJbvomsCRD4pCbFUsuABgAAC/IIAI3LD89Fn9aJu/+eNsJnTyJ1
7T9KQFkekAe681eMkVMUY1NDjYfcQjaw0BZqSxOrs7Tunjwxxxm4pG1ua3sDp99a
NiB2tJN6AOKWXfs6zg3d8igskANv1ArmKqEiUyL69O8eBO0fz2dfUw67JazWu6HE
+MYpurRph8w5Sz9Ay3STntsFngGEgB87P/UMFFioY1KebJpBNMhuGa6SrT8kxNif
ERQachtjnsZiPQddPo2AJYFuN4XxbHpRvi+N8F8T2gQIjP9Ux7muegUI3qU9q9PU
VaefYa8rHJpw3VIt+1qf0RoiW53zJD+dYhSwTH4MBeagyDOjmQiLbXRI4Ofbc1s=
=JinU
-----END PGP SIGNATURE-----
```

## The "smart" Protocol

HTTP clients that support the "smart" protocol (or both the "smart" and "dumb" protocols) MUST discover bundles by making a parameterized request for the bundle.

The request MUST contain exactly one query parameter, `service=$servicename`, where `$servicename` MUST be the service name the client wishes to contact to complete the operation. The request MUST NOT contain additional query parameters.

```bash
GET $REPO_URL?service=cnab HTTP/1.1
```

Example smart server reply:

```bash
200 OK
Content-Type: application/x-cnab-advertisement
```

If the server does not recognize the requested service name, or the requested service name has been disabled by the server administrator, the server MUST respond with the `403 Forbidden` HTTP status code.

Otherwise, smart servers MUST respond with the smart server reply format for the requested service name.

Cache-Control headers SHOULD be used to disable caching of the returned entity.

The Content-Type MUST be `application/x-$servicename-advertisement`. Clients SHOULD fall back to the dumb protocol if another content type is returned. When falling back to the dumb protocol, clients SHOULD NOT make an additional request to `$REPO_URL`, but instead SHOULD use the response already in hand. Clients MUST NOT continue if they do not support the dumb protocol.

Clients MUST validate the status code is between 200-399.

Further content negotiation and the communication protocol between the client and the server is left up entirely to the custom reply format for the requested service name.

### Searching

HTTP clients that support the "smart" protocol (or both the "smart" and "dumb" protocols) MAY search for bundles by making a request to `/search`.

HTTP servers that only support the "dumb" protocol DO NOT need to implement a search API.

The `/search` API is a global resource, used to search across an entire hostname for bundles.

Using the previous URL example, "smart" clients send requests to the following endpoint:

    https://<host>:<port>/search?q=<searchpart>

The request MUST contain two query parameters:

- `q=$query`, where `$query` MUST be a string of keywords the client wishes to use to search across the repository
- `service=$servicename`, where `$servicename` MUST be the service name the client wishes to contact to complete the operation.

The request MUST NOT contain additional query parameters.

```bash
GET https://<host>:<port>/search?servicename=cnab&q=helloworld HTTP/1.1
```

Example smart server reply:

```bash
200 OK
Content-Type: application/x-cnab-searchresults
Link: <https://<host>:<port>/search?q=helloworld&page=2>; rel="next", <https://<host>:<port>/search?q=helloworld&page=50>; rel="last"

{
  "apiVersion": "v1",
  "bundles": {
    "bacongobbler/helloworld": {
      "v2.0.0": "/v2/bacongobbler/helloworld/manifests/v2.0.0",
      "v1.0.0": "/v2/bacongobbler/helloworld/manifests/v1.0.0"
    },
    "radu-matei/helloworld": {
      "v2.0.0": "/v2/radu-matei/helloworld/manifests/v2.0.0",
      "v1.0.0": "/v2/radu-matei/helloworld/manifests/v1.0.0"
    },
  }
}
```

The Content-Type MUST be `application/x-$servicename-searchresults`. Clients MUST NOT continue if another Content-Type is returned.

The response is paginated based on the number of bundles, however the default number of pages and the number of entries per page are left up to the server.

In the above example, there are 4 entries; one for each version and its respective link.

The Link header includes pagination information. It's important for clients to form calls using Link header values instead of constructing your own URLs.

The possible `rel` values are:

| Name  | Description                                                   |
|-------|---------------------------------------------------------------|
| next  | The link relation for the immediate next page of results.     |
| last  | The link relation for the last page of results.               |
| first | The link relation for the first page of results.              |
| prev  | The link relation for the immediate previous page of results. |

Clients can traverse through the paginated response by adding another query parameter, `page=$pagenumber`, where `$pagenumber` MUST be an integer between 1 and the value in "pages".

If no `page` query parameter is set, the response MUST be the first page.

Pages are one-indexed, such that `page=1` is the FIRST page. `page=0` is NOT a valid page number.

HTTP clients MAY also add a third query parameter, `per_page=$numentries`, where `$numentries` is the number of entries the client wishes to view in a single page. Servers MAY choose to ignore this query parameter, and clients should be prepared for that.

## Motivation

Duffle repositories are a centralized location where packaged bundles can be stored and shared. They can be created by anyone to distribute their own bundles, and users can use these repositories to share, collaborate and consume bundles created by the community. It makes searching, fetching, and sharing bundles easier, secure, and manageable for both the producer and the consumer of these bundles.

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

[bundle.json]: 101-bundle-json.md
