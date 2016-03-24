import os
import subprocess

from crossrunner.collect import collect_testlibs


def prepare(config_dict, testdir, server_match, client_match):
  libs, libs2 = collect_testlibs(config_dict, server_match, client_match)
  libs.extend(libs2)

  def prepares():
    for lib in libs:
      pre = lib.get('prepare')
      if pre:
        yield pre, lib['workdir']

  def files():
    for lib in libs:
      workdir = os.path.join(testdir, lib['workdir'])
      for c in lib['command']:
        if not c.startswith('-'):
          p = os.path.join(workdir, c)
          if not os.path.exists(p):
            yield os.path.split(p)

  def make(p):
    d, f = p
    with open(os.devnull, 'w') as devnull:
      return subprocess.Popen(['make', f], cwd=d, stderr=devnull)

  for pre, d in prepares():
    subprocess.Popen(pre, cwd=d).wait()

  for p in list(map(make, set(files()))):
    p.wait()
  return True
