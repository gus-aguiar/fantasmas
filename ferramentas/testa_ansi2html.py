import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ansi2html import converte


class T(unittest.TestCase):
    def test_cor_24_bits(self):
        self.assertIn("color:#b996ff", converte("\x1b[38;2;185;150;255mx\x1b[0m"))

    def test_fundo(self):
        self.assertIn("background:#3a2a55", converte("\x1b[48;2;58;42;85mx\x1b[0m"))

    def test_reset(self):
        h = converte("\x1b[38;2;1;2;3mx\x1b[0my")
        self.assertIn("x</span>y", h)

    def test_html_escapa(self):
        self.assertIn("&lt;b&gt;", converte("<b>"))


if __name__ == "__main__":
    unittest.main()
