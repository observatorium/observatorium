{
  selectorsToLabels(labelset):: {
    [s[0]]: std.strReplace(s[1], '"', '')
    for s in [
      std.split(s, '=')
      for s in labelset
    ]
  },
}
