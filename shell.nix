with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    air
    go
    golint
    gopls
    flyctl
    golangci-lint
    sqlc
  ];
}
