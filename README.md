# stupid-caldaia

Caldaia means boiler.

Since my boiler is old and stupid, I decided to make it smart, but not too much.

Plus one can learn go, which is pretty cool.

# Interface

# Flash raspberry

Insert sd card and run and bypass windows to connect USB to WSL2 (you have to build the kernel with SD and USB drivers first)

```powershell
usbip list -l
```

```powershell
usbipd attach --wsl --busid 1-1
```

```bash
lsblk

>>> OK
sdd      8:48   1  59.5G  0 disk
├─sdd1   8:49   1   256M  0 part
└─sdd2   8:50   1  59.2G  0 part
```

```bash
sudo rpi-imager
```
