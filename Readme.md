# nrversions

## Notes

1. Update all versioned test package.json files with:
   1. A top-level `target` key that indicates the module instrumentation being
   verified
   2. A `supported: false` key+value on any test descriptors that are testing
   _unsupported_ versions, see the `elastic` test descriptor for 7.13.0
