#!/usr/bin/env python3

# Tests for `fission-workflows status`
import sys, os
sys.path.append(os.path.dirname(__file__)) # Needed to import testutils when invoked from another wd.
import testutils

# Check for correctly deployed cluster
testutils.run(["fission-workflows", "status"], check=True)

# Check if hitting a non-existing cluster results in an error
out = testutils.run(["fission-workflows", "--url", "http://127.0.0.1:1337", "status"])
if out.returncode == 0:
    raise "%s should have failed!" % " ".join(out.args)