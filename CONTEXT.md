lint ci failed on install and run golangci-lint step. some output below:

  Checking for go.mod: go.mod
  Received 138412032 of 175392406 (78.9%), 132.0 MBs/sec
  Cache Size: ~167 MB (175392406 B)
  /usr/bin/tar -xf /home/runner/work/_temp/f73cb1a6-38ae-4e9c-b813-24ac250b5a55/cache.tzst -P -C /home/runner/work/architect/architect --use-compress-program unzstd
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/LICENSE: Cannot open: File exists
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/CONTRIBUTING.md: Cannot open: File exists
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/README.md: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/endian_little.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_mips64x.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_linux_noinit.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_arm.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_darwin_x86.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_other_arm64.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_wasm.go: Cannot open: File exists
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_arm64.s: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_linux.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_ppc64x.go: Cannot open: File exists
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/asm_aix_ppc64.s: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_test.go: Cannot open: File exists
  Error: /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/runtime_auxv_go121.go: Cannot open: File exists
  /usr/bin/tar: ../../../go/pkg/mod/golang.org/x/sys@v0.31.0/cpu/cpu_s390x.s: Cannot open: File exists

...

  Running [/home/runner/golangci-lint-1.64.8-linux-amd64/golangci-lint run --out-format=github-actions --timeout=5m] in [] ...
  Error: ineffectual assignment to path (ineffassign)

  level=warning msg="[config_reader] The output format `github-actions` is deprecated, please use `colored-line-number`"

  Error: issues found
  Ran golangci-lint in 1027ms

