import os
from itertools import permutations
import subprocess
import tempfile
import sys

RULES_PATH = './rules'
DEFAULT_FILE = '_default.yml'
DOCKER_COMMAND = ['docker', 'run', '-v', 'REPLACEME:/etc/falco/falco_rules.yaml', '-e', 'SKIP_DRIVER_LOADER=1', 'falcosecurity/falco', '/usr/bin/falco', '-V', '/etc/falco/falco_rules.yaml']

files = []
for path in os.listdir(RULES_PATH):
    if path == DEFAULT_FILE:
        continue
    if os.path.isfile(os.path.join(RULES_PATH, path)):
        files.append(os.path.join(RULES_PATH, path))

# modify default to be the full path
DEFAULT_FILE = os.path.join(RULES_PATH, DEFAULT_FILE)

for l in range(1,len(files)+1):
    perm = permutations(files, l)
    for p in list(perm):
        test = [DEFAULT_FILE,] + list(p)
        print(f'Running test on files in order: {test}')
        # create a temp file with the concated files
        tmp = tempfile.NamedTemporaryFile(delete=False)
        with open(tmp.name, 'w') as f:
            for fname in test:
                with open(fname) as infile:
                    for line in infile:
                        f.write(line)
                    # write an additional newline because that's how the 
                    # go API will concat the files
                    f.write('\n')
        # run the test
        dcommand = DOCKER_COMMAND.copy()
        dcommand[3] = dcommand[3].replace("REPLACEME", tmp.name)
        print(dcommand)
        res = subprocess.call(dcommand)
        if res != 0:
            print(f'Rule failed: {test}')
            sys.exit(1)
        