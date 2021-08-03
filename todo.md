## templating language
- make it so, that dynamic values can reference other dynamic values
  - example ```
  {
    "0": "hello world",
    "1": "0"
  }```
  - this way sending the same value multiple times within one patch will be prevented
  - this dependency should only be sent **once**! with the big payload upon mount. In patches value `"1"` can be derived from `"0"`