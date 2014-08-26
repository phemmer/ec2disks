ec2disks is a utility for discovering information about volumes attached to an Amazon EC2 instance.
It can output the information in an easily viewable format, similar to `blkid`, or it can be used in udev rules to set up symlinks.

It retrieves this information from the EC2 instance meta-data (http://169.254.169.254/), as well as API calls to the EC2 API. For the API calls to work, you either need to use IAM instance roles, or you need to set the environment variables `AWS_ACCESS_KEY_ID`/`AWS_ACCESS_KEY` & `AWS_SECRET_ACCESS_KEY`/`AWS_SECRET_KEY`.

## Example usage:

    $ ec2disks
    /dev/xvda: TYPE=ebs ID=vol-b4287bb4 ALIASES=disk/ec2/ami,disk/ec2/root,disk/ec2/vol-b4287bb4
    /dev/xvdb: TYPE=ephemeral ID= ALIASES=disk/ec2/ephemeral0
    /dev/xvdf: TYPE=ebs ID=vol-d15201d1 ALIASES=disk/ec2/vol-d15201d1

    $ ec2disks /dev/xvdb
    /dev/xvdb: TYPE=ephemeral ID= ALIASES=disk/ec2/ephemeral0

    $ ec2disks -A /dev/xvda
    disk/ec2/ami disk/ec2/root disk/ec2/vol-b4287bb4

    $ ec2disks -s -A /dev/xvda
    ami root vol-b4287bb4

## Udev rule
As mentioned, you can use the utility to set up udev symlinks. The links will be placed in `/dev/disks/ec2/`.

Setting this up is very simple.  
After the utility is installed on the system, create `/etc/udev/rules.d/10-ec2disks.rules` as:

    SUBSYSTEM=="block", SUBSYSTEMS=="xen", ATTR{capability}=="*", PROGRAM="ec2disks -A %k", SYMLINK+="%c"

*Note: Depending on where you installed the utility, you may need to fully qualify the path. It's safest to do this anyway, whether needed or not.*

To make the rules take effect without a reboot, simply issue:

    udevadm trigger --subsystem-match=block --action=change

&nbsp;

    $ ls -l /dev/disk/ec2/
    total 0
    lrwxrwxrwx 1 root root 10 Aug 26 16:04 ami -> ../../xvda
    lrwxrwxrwx 1 root root 10 Aug 26 16:04 ephemeral0 -> ../../xvdb
    lrwxrwxrwx 1 root root 10 Aug 26 16:04 root -> ../../xvda
    lrwxrwxrwx 1 root root 10 Aug 26 16:04 vol-b4287bb4 -> ../../xvda
    lrwxrwxrwx 1 root root 10 Aug 26 16:41 vol-d15201d1 -> ../../xvdf
