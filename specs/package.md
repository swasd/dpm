# Cloud Native Package

# Package Format
A cloud native package managed by `dpm` is inspired by package management systems for Linux distributions.
Unlike an application package, a cloud native package will be installed on a pool of resources formed by multi-cloud providers.
So it is a package to install on multi data centers as a single machine. Major components of the package are `provision` and `composition`.

The content of a _package_ is a `tar` file consists of

  0. SPEC.yml as the first entry.
  0. A provision file, described in `SPEC.yml`.
  0. A composition file, described in `SPEC.yml`.
  0. A list of package dependencies, under their SHA256 hash sub-directories.

# SPEC.yml
