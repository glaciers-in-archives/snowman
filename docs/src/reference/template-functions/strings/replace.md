# `replace`

`replace` function replaces substrings in a string with a given string.

The first argument is the string to which the replacement is to be applied, the second argument is the substring to be replaced, the third argument is the substring to replace with and finally the fourth argument is the number of replacements to be made. If the fourth argument is is set to `-1` then all occurrences will be replaced.

```
{{ replace "hello world" "world" "earth" -1 }}
```
