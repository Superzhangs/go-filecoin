#!/usr/bin/env bash

install_precompiled() {
  echo "precompiled libfil_secp256k1 not supported yet"
  return 1
}

install_local() {
  if ! [ -x "$(command -v cargo)" ] ; then
    echo 'Error: cargo is not installed.'
    echo 'Install Rust toolchain to resolve this problem.'
    exit 1
  fi

  git submodule update --init --recursive

  pushd crypto/rust-fil-secp256k1

  cargo --version
  cargo update
  cargo build --release

  popd

  mkdir -p crypto/include
  mkdir -p crypto/lib/pkgconfig

  cp crypto/rust-fil-secp256k1/target/release/libfil_secp256k1.h ./crypto/include/
  cp crypto/rust-fil-secp256k1/target/release/libfil_secp256k1.a ./crypto/lib/
  cp crypto/rust-fil-secp256k1/target/release/libfil_secp256k1.pc ./crypto/lib/pkgconfig/
}

if [ -z "$FILECOIN_USE_PRECOMPILED_LIBSECP256K1_SIGNATURES" ]; then
  echo "using local libfil_secp256k1"
  install_local
else
  echo "using precompiled libfil_secp256k1"
  install_precompiled

  if [ $? -ne "0" ]; then
    echo "failed to find or obtain precompiled libfil_secp256k1, falling back to local"
    install_local
  fi
fi
