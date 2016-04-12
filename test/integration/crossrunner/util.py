import copy


def merge_dict(base, update):
  """Update dict concatenating list values"""
  res = copy.deepcopy(base)
  for k, v in list(update.items()):
    if k in list(res.keys()) and isinstance(v, list):
      res[k].extend(v)
    else:
      res[k] = v
  return res
