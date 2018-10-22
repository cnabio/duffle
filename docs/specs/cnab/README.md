# Cloud Native Application Bundles (CNAB)

## CNAB Package Specification

This document describes the CNAB (Cloud Native Application Bundles) package format.

A CNAB bundle consists of:
- A [bundle.json file]
- An [invocation image]

Bundles can be packaged in [thin] or [thick] formats, and can be stored and served from [bundle repositories].

## Table of Contents


## Conventions Used in This Document

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in BCP 14 [RFC2119](https://www.rfc-editor.org/info/rfc2119) and [RFC8174](https://www.rfc-editor.org/info/rfc8174) when, and only when, they appear in all capitals, as shown here.