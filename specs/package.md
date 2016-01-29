# Cloud Native Package

# Package Format
A cloud native package managed by `dpm` is named after package management systems for Linux distributions.
A _package_ is a `tar` file consists of

  0. SPEC.yml as the first entry.
  0. A provision file, described in `SPEC.yml`.
  0. A composition file, described in `SPEC.yml`.
  0. A list of package dependencies, under their SHA256 hash sub-directories.

# SPEC.yml
