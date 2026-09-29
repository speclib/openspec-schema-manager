package ansi

import "testing"

func TestStrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain text is untouched", in: "Registry", want: "Registry"},
		{name: "empty", in: "", want: ""},
		{name: "bold", in: "\x1b[1mRegistry\x1b[m", want: "Registry"},
		{name: "per character underline", in: "\x1b[4;4mR\x1b[m\x1b[4;4me\x1b[m", want: "Re"},
		{name: "colour", in: "\x1b[38;5;214mwarn\x1b[0m", want: "warn"},
		{name: "cursor movement", in: "a\x1b[2Ab", want: "ab"},
		{name: "alternate screen", in: "\x1b[?1049hhi\x1b[?1049l", want: "hi"},
		{name: "an OSC ending in BEL", in: "\x1b]0;title\x07body", want: "body"},
		{name: "an OSC ending in ST", in: "\x1b]8;;https://example.test\x1b\\link", want: "link"},
		{name: "a lone escape at the end", in: "done\x1b", want: "done"},
		{name: "an unterminated CSI", in: "x\x1b[38;5", want: "x"},
		{name: "a two byte escape", in: "a\x1b(Bb", want: "ab"},
		{name: "newlines survive", in: "\x1b[1ma\x1b[m\nb", want: "a\nb"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := Strip(tc.in); got != tc.want {
				t.Errorf("Strip(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
