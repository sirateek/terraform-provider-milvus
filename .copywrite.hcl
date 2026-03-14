project {
  license = "MIT"

  # (OPTIONAL) Represents the copyright holder used in all statements
  # Default: IBM Corp.
  copyright_holder = "Siratee K."

  # (OPTIONAL) Represents the year that the project initially began
  # This is used as the starting year in copyright statements
  # If set and different from current year, headers will show: "copyright_year, year-2"
  # If set and same as year-2, headers will show: "copyright_year"
  # If not set (0), the tool will auto-detect from git history (first commit year)
  # If auto-detection fails, it will fallback to current year only
  # Default: 0 (auto-detect)
  # copyright_year = 0

  # (OPTIONAL) A list of globs that should not have copyright or license headers .
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  # Default: []
  header_ignore = [
    # "vendor/**",
    # "**autogen**",
    ".idea/**"
  ]

  # (OPTIONAL) Links to an upstream repo for determining repo relationships
  # This is for special cases and should not normally be set.
  # Default: ""
  # upstream = "hashicorp/<REPONAME>"
}