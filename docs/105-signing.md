# Signing, Attesting, Verifying, and Validating CNAB Bundles

Package signing is a method frequently used by package managers as a way of ensuring, to the user, that the package was _vetted by a trusted entity_ and _has since, not been altered_. Signing is a critical piece of infrastructure security.

This document outlines how CNAB bundles use a multi-layered fingerprinting and signing strategy to meet these criteria.

## Summary

- Every image and invocation image field in the `bundle.json` must have a `digest:` field that must contain a digest of the image.
- Digests are computed in accordance with the underlying image type (e.g. OCI bundles are validated by computing the top hash of a Merkle tree, VM images are computed by digest of the image)
- Signed bundles are clear-signed `bundle.json` files according to the Open PGP specification. When present, these are used in lieu of the unsigned `bundle.json` file.
- Authority is granted by the signed bundle, and integrity is granted via the image digests embedded in the bundle.json
- Attestations provide a mechanism for making additional guarantees about a bundle. Attesting a bundle may indicate that a release has been certified, or passed tests, or manual checked. It is a method to attach additional cryptographically based assurances to a bundle

## Image Integrity with Digests

CNAB correlates a number of images, of varying types, together in a single bundle. This section of the specification defines how image integrity is to be tested via digests and checksumming.

### Digests, OCI CAS, and Check Summing

A frequently used tool for validating that a file has not been changed between time T and time T+n is _checksumming_. In this strategy, the creator of the file runs a cryptographic hashing function (such as a SHA-512) on a file, and generates a _digest_ of the file's content. The digest can then be transmitted to a recipient separately from the file. The recipient can then re-run the same cryptographic hashing function on the file, and verify that the two functions are identical. There are more elaborate strategies for digesting complex objects, such as a [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree), which is what the OCI specification uses. In any event, the output _digest_ can be used to later verify the integrity of an object.

The OCI specification contains a [standard for representing digests](https://github.com/opencontainers/image-spec/blob/master/descriptor.md#digests). In its simplest case, it looks like this:

```text
ALGO:DIGEST
```

Where `ALGO` is the name of the cryptographic hashing function (`sha512`, `md5`, `blake2`...) plus some optional metadata, and DIGEST is the ASCII representation of the hash (typically as a hexadecimal number).

> Note: The OCI specification only allows `sha256` and `sha512`. This is not a restriction we make here.

For example:

```text
sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b
```

### Digesting Objects in the `bundle.json`

CNAB is composed of a `bundle.json` and a number of supporting images. Those images are referenced by the `bundle.json`. Thus, digesting those artifacts and including their digest in the `bundle.json` provides a convenient way to store (and locate) digests.

To that end, anything that shows up in the `invocationImages` or `images` section of the `bundle.json` _must_ have a `digest`:

```json
{
    "name": "helloworld",
    "version": "0.1.2",
    "invocationImages": [
        {
            "imageType": "docker",
            "image": "technosophos/helloworld:0.1.0",
            "digest": "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
        }
    ],
    "images": [
        {
            "name": "image1",
            "digest": "sha256:aaaa624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b",
            "uri": "urn:image1uri",
            "refs": [
                {
                    "path": "image1path",
                    "field": "image.1.field"
                }
            ],
        }
    ]
}
```

Objects must contain a `digest` field even if the digest is present in another field. This is done to provide _access uniformity_.

> OCI images, for example, may embed a digest in the image's _version_ field. According to this specification, while this is allowed, it does not remove the requirement that the `digest` field be present and filled.

Different formats (viz. OCI) provide definitions for validating a digest. When possible, images should be validated using these definitions, according to their `imageType`. If a particular image type does not already define what it means to have a digest verified, the default method is to retrieve the object as-is, and checksum it in the format in which it was delivered when accessed.

Drivers may choose to accept the digesting by another trusted agent in lieu of performing the digest algorithm themselves. For example, if a driver requests that a remote agent install an image on its behalf, it may trust that the image digest given by that remote agent is indeed the digest of the object in question. And it may then compare that digest to the `bundle.json`'s digest. In such cases, a driver _should_ ensure that the channel between the driver itself and the trusted remote agent is itself secured (for example, via TLS). Failure to do so will invalidate the integrity of the check.

## Signing the `bundle.json`

The `bundle.json` file will contain the digests of all executable objects. That is, everything in the `invocationImages` and `images` sections will have digests that will make it possible to ensure that their content has not been tampered with.

Consequently, the `bundle.json` acts as an authoritative resource for image integrity. To act as an authoritative source, however, it must provide an additional assertion: The `bundle.json` must assert the intention of the bundle creator, in marking this as a _verified bundle_.

This is accomplished by _signing the bundle_.

The signature method used by CNAB is defined by the [Open PGP standard](https://tools.ietf.org/html/rfc4880)'s digital signatures specification. In short, a _packaging authority_ (the individual responsible for packaging or guaranteeing the package), signs the bundle with a _private key_. The packaging authority distributes the accompanying public key via other channels (not specified herein, but including trusted HTTP servers, Keybase, etc.)

The _package recipient_ (the consumer of the package) may then retrieve the public keys. Upon fetching a signed bundle, the package recipient may then _verify_ the signature on the bundle by testing it against the public key.

An Open PGP signature follows [the format in Section 7 of the specification](https://tools.ietf.org/html/rfc4880#section-7):

```text
-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA512

   <BODY>

-----BEGIN PGP SIGNATURE-----
Comment: <GENERATOR>

<SIGNATURE>
-----END PGP SIGNATURE-----
```

In the above, `<BODY>` is the entire contents of the `bundle.json`, `<GENERATOR>` is the optional name of the program that generated the signature, and `<SIGNATURE>` is the signature itself.

For example, here is a `bundle.json`:

```json
{
    "name": "foo",
    "version": "1.0",
    "invocationImages": [
        {
            "imageType": "docker",
            "image": "technosophos/helloworld:0.1.2",
            "digest": "sha256:aca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120685"
        }
    ],
    "images": [],
    "parameters": {},
    "credentials": {}
}
```

This is signed using the technique called _clear signing_ (OpenPGP, Section 7), which preserves the input along with the cryptographic signature:

```text
-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA512

{
    "name": "foo",
    "version": "1.0",
    "invocationImages": [
        {
            "imageType": "docker",
            "image": "technosophos/helloworld:0.1.2",
            "digest": "sha256:aca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120685"
        }
    ],
    "images": [],
    "parameters": {},
    "credentials": {}
}
-----BEGIN PGP SIGNATURE-----
Comment: GPGTools - https://gpgtools.org

iQEzBAEBCgAdFiEE+yumTPTtBoSSRd/j3NX15e8yw0UFAluEHv0ACgkQ3NX15e8y
w0UyKQf/Tb/mURLiHWmw68q7cjAHx7wVgjClo34oB07uY1RCvjMiK4sXaoKC+0KO
pQOC/15HY9f2aazPHig//aqNpFyyHcpX9XjvH51CbXiNcFvIv7JgmFwr7WIUY7cS
FsaFSCS53Z5HqCQ/SYB9OU4R+uwBW/gKmP7vBGieNkEhqHQklQG9vc9zUQjuTlYp
KIW9cGd0rKWzs8wwiW9FytIM43x54sHmtXRnWxu6zNReXr32u6ZFPrfVA0yoAJQ4
7iDhcM/VqL4j0xxfFmZuqkRCtsbgD6iUmL8VzINODGsF4lHFQrl2sKXAqMoIXyCw
ANjudClHNUNQFojriAX8YAO4V2yGVg==
=OoBW
-----END PGP SIGNATURE-----
```

Note that because we have _clear signed_ the `bundle.json`, there is no longer any need to transmit the `bundle.json` separately from the signed bundle. In fact, due to encoding differences, it is _preferable to use the signed bundle in lieu of the unsigned bundle_. 

## Attestations

The purpose of a clear-signed signature is to assert that a particular artifact (the bundle) has been generated by an authority, where authority is marked by a cryptographic signature.

Attestations provide a way to add multiple assurances to a bundle.

An attestation is used as a placeholder for the statement "The signing party has certified (attested) that condition X has been met". For example, when certification steps need to occur for a given bundle, one may use attestations to prove that the certification has been performed. In this case, the certifying party performs the process, and then (upon the bundle's passing), the certifying party adds a _detached signature_ to the set of signatures associated with the bundle.

In a more complex case, a signed package may be certified by one party for use in one way (we'll call this "certification A"). A different party may certify the bundle for a different case ("certification B"). Note that in this case, the certifications are _independent_, and are presumably done with separate justifications. ("certificate A ensures this bundle is suitable for internal use", "certificate B ensures this bundle is suitable for use by our partners".) Because of this feature, attestations _are not chainlike_. Each individual attestation must be verifiable without reference to any other attestations (including the original clear-signed signature).

Detached signatures are described in [section 11.4](https://tools.ietf.org/html/rfc4880#section-11.4) of the OpenPGP specification. Attestations are to be performed by extracting the `bundle.json` from a signed bundle, and then signing that same text object. A verification of a detached signature should use the `bundle.json` text as a basis for its verification. A bundle is considered _attested_ (or _attestation verified_) when a the bundle verification passes for the expected key used in that attestation.

They key used for an attestation must be used _only for that attestation_. Implementations _may_ use a subkey (of a master key) for specific attestations, while preserving other subkeys to perform other attestations or signings.

Implementations of CNAB _ought_ to support creating and verifying attestations. Implementations of CNAB _may_ support favoring an attestation with equal weight as the original signature. Attestations _may_ be stored with the signed bundle, though there is no requirement that attestations be stored in a specific place.

Attestations _may_ use the `COMMENT:` field of a detached signature to indicate, in a human-friendly way, what the attestation is for. However, agents _must not_ consider this information definitive. Comment fields are not calculated into the signature and can be easily modified. Instead, attestation _must_ be based solely on the key.

```text
-----BEGIN PGP SIGNATURE-----
Comment: Attestation - Certified for Internal Use

iQEzBAABCgAdFiEE+yumTPTtBoSSRd/j3NX15e8yw0UFAluayWsACgkQ3NX15e8y
w0VvVQgA1FtF03jqQgiAkxd707ELtmrKX5dcfIYtEr3o5fBGtckUebV5RYFwfQqZ
fYoTVEiAzgtR6ceXQB+SjCj8KD5uhf2nzX5eKIAmhyCLKibVBVCaTlTsKzNR/Xe4
fPWp/nSlNo6Xc2kwx6RRPPMpYk/7WhXm7iIl7MmHmveHmTM1oTdrzhf/y1ZTc0Vu
qdBSRvsDJMnaf+iB2g9r113ee12UBta9pbLIXjlWpv4PknL7QNsp2B0KeExQXgvZ
2KKrz+ndWr3I5aONa6Zr9hh3NdZc/oa1peqJaCJtsrLj08/+WiwdWTWG3/8k+toW
UkmNdIkOChvHv42XkWnF1t1Hyi51ig==
=sQ9B
-----END PGP SIGNATURE-----
```

## Key Revocation and Expiration

When public keys are expired or revoked, bundles signed with those keys become invalid.
They must be re-signed with a valid key.

CNAB verification tools _should_ handle the key revocation case.

## TODO

- Do we want to allow encrypted `bundle.json` files? This is actually trivially easy (it's another option on Open PGP)
- There's a "chicken and egg" problem if we try to store the signed bundle.json inside of the CNAB invocation image. My preference is to always require the signed bundle to be external of the image. It's basically useless if it is inside the image anyway, because the image must be transmitted and extracted before the signature can be verified, which invalidates the verification step.
- We need to specify this format for dense bundles

Next section: [declarative images](106-declarative-images.md)
