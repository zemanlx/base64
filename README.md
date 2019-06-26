# base64

Backward compatible alternative of Linux `base64` plus

-   URL encoding option
-   No padding option (both for standard and URL encoding)


    Usage: base64 [OPTION]... [FILE]

    Base64 encode or decode FILE, or standard input, to standard output.
    With no FILE, or when FILE is -, read standard input.

      -d, --decode           decode data
      -h, --help             print this help
      -i, --ignore-garbage   when decoding, ignore non-alphabet characters
      -n, --no-padding       omit padding
      -u, --url              use URL encoding according RFC4648
      -v, --version          output version information and exit
      -w, --wrap uint        wrap encoded lines after COLS character,
                             use 0 to disable line wrapping (default 76)

    The data are encoded as described for the base64 alphabet in RFC 4648.
    When decoding, the input may contain newlines in addition to the bytes of
    the formal base64 alphabet.  Use --ignore-garbage to attempt to recover
    from any other non-alphabet bytes in the encoded stream.
