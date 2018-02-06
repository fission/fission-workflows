import subprocess, sys

def run(cmd, input=None, timeout=60, check=False):
    print("--- CMD: '%s' ---" % " ".join(cmd))
    result = subprocess.run(cmd,
                            input=input,
                            stdout=sys.stdout,
                            # Combine stderr and stdout
                            stderr=subprocess.STDOUT,
                            # Timeout in seconds
                            timeout=timeout,
                            encoding='ascii')
    status = "FAILED"
    if result.returncode == 0:
        status = "SUCCESS"
    print("\n--- [%s] CMD: '%s' (return code: %d) ---" % (status, " ".join(cmd), result.returncode))
    if check:
        # This throws a subprocess.CalledProcessError for a non-zero exit status
        result.check_returncode()
    return result