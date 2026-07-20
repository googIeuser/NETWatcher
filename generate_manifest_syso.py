#!/usr/bin/env python3
from __future__ import annotations

import struct
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
OUT = ROOT / "app_manifest_windows_amd64.syso"

manifest = b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity version="2.2.2.0" processorArchitecture="amd64" name="NetWatcher.NetWatcher" type="win32"/>
  <description>NetWatcher Internet Connection Monitor</description>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
    <security>
      <requestedPrivileges>
        <requestedExecutionLevel level="asInvoker" uiAccess="false"/>
      </requestedPrivileges>
    </security>
  </trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
    <application>
      <supportedOS Id="{e2011457-1546-43c5-a5fe-008deee3d3f0}"/>
      <supportedOS Id="{35138b9a-5d96-4fbd-8e2d-a2440225f93a}"/>
      <supportedOS Id="{4a2f28e3-53b9-4441-ba9c-d69d4a4a6e38}"/>
      <supportedOS Id="{1f676c76-80e1-4239-95bb-83d0f6d0da78}"/>
      <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
    </application>
  </compatibility>
</assembly>
'''


def align(value: int, boundary: int) -> int:
    return (value + boundary - 1) & ~(boundary - 1)


def section_symbol(name: bytes, section_number: int, size: int, reloc_count: int) -> bytes:
    assert len(name) <= 8
    symbol = struct.pack("<8sIhHBB", name.ljust(8, b"\0"), 0, section_number, 0, 3, 1)
    # IMAGE_AUX_SYMBOL_SECTION: Length, NumberOfRelocations, NumberOfLinenumbers,
    # CheckSum, Number, Selection, reserved, HighNumber.
    aux = struct.pack("<IHHIhBBH", size, reloc_count, 0, 0, 0, 0, 0, 0)
    assert len(symbol) == 18 and len(aux) == 18
    return symbol + aux


# IMAGE_RESOURCE_DIRECTORY tree: RT_MANIFEST (24) -> ID 1 -> en-US (0x409)
rsrc1 = bytearray()
rsrc1 += struct.pack("<IIHHHH", 0, 0, 0, 0, 0, 1)
rsrc1 += struct.pack("<II", 24, 0x80000018)
rsrc1 += struct.pack("<IIHHHH", 0, 0, 0, 0, 0, 1)
rsrc1 += struct.pack("<II", 1, 0x80000030)
rsrc1 += struct.pack("<IIHHHH", 0, 0, 0, 0, 0, 1)
rsrc1 += struct.pack("<II", 0x0409, 0x00000048)
# OffsetToData is relocated by IMAGE_REL_AMD64_ADDR32NB to $R000000.
rsrc1 += struct.pack("<IIII", 0, len(manifest), 0, 0)
assert len(rsrc1) == 0x58

rsrc2 = manifest + (b"\0" * (align(len(manifest), 4) - len(manifest)))

coff_header_size = 20
section_table_size = 40 * 2
rsrc1_offset = coff_header_size + section_table_size
reloc_offset = rsrc1_offset + len(rsrc1)
rsrc2_offset = align(reloc_offset + 10, 4)
symbol_table_offset = rsrc2_offset + len(rsrc2)

header = struct.pack(
    "<HHIIIHH",
    0x8664,  # AMD64
    2,
    int(time.time()),
    symbol_table_offset,
    6,
    0,
    0,
)

characteristics = 0x40000040  # initialized data | readable
section1 = struct.pack(
    "<8sIIIIIIHHI",
    b".rsrc$01",
    0,
    0,
    len(rsrc1),
    rsrc1_offset,
    reloc_offset,
    0,
    1,
    0,
    characteristics,
)
section2 = struct.pack(
    "<8sIIIIIIHHI",
    b".rsrc$02",
    0,
    0,
    len(rsrc2),
    rsrc2_offset,
    0,
    0,
    0,
    0,
    characteristics,
)

relocation = struct.pack("<IIH", 0x48, 5, 0x0003)  # IMAGE_REL_AMD64_ADDR32NB

symbols = bytearray()
symbols += struct.pack("<8sIhHBB", b"@feat.00", 0x11, -1, 0, 3, 0)
symbols += section_symbol(b".rsrc$01", 1, len(rsrc1), 1)
symbols += section_symbol(b".rsrc$02", 2, len(rsrc2), 0)
symbols += struct.pack("<8sIhHBB", b"$R000000", 0, 2, 0, 3, 0)
assert len(symbols) == 6 * 18

blob = bytearray(header + section1 + section2)
assert len(blob) == rsrc1_offset
blob += rsrc1
assert len(blob) == reloc_offset
blob += relocation
blob += b"\0" * (rsrc2_offset - len(blob))
blob += rsrc2
assert len(blob) == symbol_table_offset
blob += symbols
blob += struct.pack("<I", 4)  # empty COFF string table

OUT.write_bytes(blob)
print(f"wrote {OUT} ({len(blob)} bytes); manifest={len(manifest)} bytes")
