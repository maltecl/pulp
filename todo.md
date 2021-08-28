## templating language
- make it so, that dynamic values can reference other dynamic values
  - example ```{
    "0": "hello world",
    "1": "0"}```
  - this way sending the same value multiple times within one patch will be prevented
  - this dependency should only be sent **once**! with the big payload upon mount. In patches value `"1"` can be derived from `"0"`



### Json Path
  - use json path for slimmer patches


### run go import on the output file

### last token after goSource seems to be missing 
```handlebars
	<ul>
    {{ for _, line := range outLines}}
      <li> {{ line }} </li>
    {{ end }}
  </ul>
```
## If
  - make it so that when the condition is `True` the dynamic values for the `False` `StaticDynamic` are not sent. 
  - When the condition later flips, those dynamic values are sent 
