{ lib, buildGoModule, version ? "0.1.0" }:

buildGoModule {
  pname = "ossm";

  inherit version;

  src = ./.;

  subPackages = [ "cmd/ossm" ];

  ldflags = [ "-s" "-w" "-X main.version=${version}" ];

  vendorHash = null;

  doCheck = false;

  meta = {
    description = "Discover, inspect, install, author and compose OpenSpec workflow schemas";
    homepage = "https://github.com/speclib/openspec-schema-manager";
    mainProgram = "ossm";
    license = lib.licenses.mit;
  };
}
