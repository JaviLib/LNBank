# LNBank: Your Personal Bitcoin Lightning Network Node Made Easy

LNBank is a user-friendly solution for setting up your very own Bitcoin Lightning network node quickly and efficiently.

LNBank is designed as an all-encompassing package that's simple to install and configure across various platforms: Linux, MacOS, and Windows. It boasts minimal system requirements which means even older devices can handle it with ease. Plus, setting up your own Bitcoin Lightning node becomes a breeze in less than 5 minutes.

Why choose LNBank?

- No need for Docker or virtualization – LNBank runs all the services natively on your machine.
- Low resource consumption: It requires just around 5% CPU usage (1% on newer machines) and uses approximately 2 to 4 GB of RAM.
- Compatibility with desktop computers and older laptops, without the need for any additional hardware investments.
- Effortlessly multitask while your LNBank runs in the background – it's designed not to interfere with other applications or slow down your system. It minimizes itself to the status bar and stays there quietly until you need it.
- LNBank leverages Neutrino technology, which means that unlike traditional Bitcoin nodes, it doesn't require downloading the entire blockchain and will need less than 10Gb of SSD. With LNBank, you can be up-and-running in just 5 minutes – no lengthy waiting periods involved
  or expensive SSD need to be bought.

## Powered by LND and LNBits

## Requirements

1. **A computer capable of running 24/7**. LND requires constant connectivity to function effectively, receiving, sending, and routing payments. If it is offline, other peers may choose to close the channel they have with you due to locked liquidity, which is unacceptable. If you are unable to fulfill this requirement, please reconsider participating at a later time when you are prepared to assume the responsibility. An older laptop can suffice if battery degradation/condition is of no concern to you. Ensure that auto sleep or suspension is disabled in the general settings of your computer. LNBank also works great on mini-PCs and Mac-Minis, but ensure you have reliable power supply.
2. **At least 10GB of internal SSD space** is necessary. External hard drives or old HDDs are not acceptable for LNBank as the installation and execution must occur within the home directory, which should always be an internal SSD. While it's theoretically possible to trick LNBank into using an external drive, this poses unnecessary risks: accidental disconnection or connector failure could result in database corruption and the forced closure of channels. HDDs are prone to failing when they become old or heavily used, so they are not recommended for this purpose. If you suspect that your SSD may be unreliable, cease operation of your LNBank node and migrate it to another computer before a potential failure occurs.
3. **4GB of RAM** is sufficient if no additional programs will run simultaneously on your computer. If you intend to use other applications concurrently, 8GB or even 16GB is recommended for optimal performance.
4. **A multicore 64-bit processor**. Any Intel x86 or arm64 processors from the past 15 years should be sufficient.
5. Compatible operating system: **MacOS, Linux, or Windows**.
6. Special attention for Windows computers: frequent restarts interrupt node operations, malware is prevalent and may lead to financial losses, and overall system reliability is suboptimal (although improvements have been made in recent years). If you anticipate having more than 0.1 BTC in your node, it's strongly advised that you replace Windows with Linux for added security.
