import io
import unittest
import urllib.error
from unittest.mock import call, patch

import http_download


class PinnedHTTPTests(unittest.TestCase):
    def error(self, status):
        return urllib.error.HTTPError("https://example.test/pin", status,
                                      "gateway", {}, io.BytesIO(b"failure"))

    def test_gateway_retry_preserves_request_and_closes_errors(self):
        for status in (502, 503, 504):
            with self.subTest(status=status):
                errors = [self.error(status), self.error(status)]
                response = object()
                request = object()
                with patch.object(http_download.urllib.request, "urlopen",
                                  side_effect=[*errors, response]) as opened, \
                     patch.object(http_download.time, "sleep") as sleep, \
                     patch.object(http_download.sys, "stderr", io.StringIO()) as log:
                    self.assertIs(http_download.open_pinned(request, 17), response)
                    self.assertEqual(opened.call_args_list,
                                     [call(request, timeout=17)] * 3)
                    self.assertEqual(sleep.call_args_list, [call(1), call(2)])
                    self.assertIn("retry 3/3", log.getvalue())
                    self.assertTrue(all(error.closed for error in errors))

    def test_gateway_failure_remains_terminal_after_three_attempts(self):
        errors = [self.error(504) for _ in range(3)]
        with patch.object(http_download.urllib.request, "urlopen", side_effect=errors) as opened, \
             patch.object(http_download.time, "sleep") as sleep, \
             patch.object(http_download.sys, "stderr", io.StringIO()):
            with self.assertRaises(urllib.error.HTTPError) as caught:
                http_download.open_pinned(object())
            self.assertIs(caught.exception, errors[-1])
            self.assertEqual(opened.call_count, 3)
            self.assertEqual(sleep.call_count, 2)
            caught.exception.close()

    def test_other_errors_are_not_retried(self):
        for error in [self.error(403), self.error(404), self.error(429),
                      ValueError("integrity"), urllib.error.URLError("network")]:
            with self.subTest(error=error), \
                 patch.object(http_download.urllib.request, "urlopen", side_effect=error) as opened, \
                 patch.object(http_download.time, "sleep") as sleep:
                with self.assertRaises(type(error)):
                    http_download.open_pinned(object())
                self.assertEqual(opened.call_count, 1)
                sleep.assert_not_called()
                if isinstance(error, urllib.error.HTTPError):
                    error.close()
