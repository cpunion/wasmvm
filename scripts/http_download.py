"""Bound transient gateway retries before opening a pinned input stream."""

import sys
import time
import urllib.error
import urllib.request


def open_pinned(request, timeout=60):
    # Content, size, digest and extraction failures remain terminal in callers.
    # Never retry a partially consumed stream or choose another input URL.
    for attempt in range(3):
        try:
            return urllib.request.urlopen(request, timeout=timeout)
        except urllib.error.HTTPError as error:
            if error.code not in (502, 503, 504) or attempt == 2:
                raise
            error.close()
            print(f"Pinned input HTTP {error.code}; retry {attempt + 2}/3 "
                  f"after {attempt + 1}s", file=sys.stderr)
            time.sleep(attempt + 1)
