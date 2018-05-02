yum install -y lvm2  
cp lvmd/lvmd /usr/bin/
cp lvmd/lvmd.service /usr/lib/systemd/system/
systemctl enable lvmd
systemctl start lvmd

DEVS=$*
while [ -z $DEVS ]; do
    echo "Input devices/partitions used to make lvm, seperate by a space. "
    echo "Note: All devices/partitions inputed will be force umounted and wiped."
    read -p "Please input , e.g. (/dev/sdb /dev/sdc1) > " DEVS
done

echo "Force umounting $DEVS"
umount -l $DEVS

echo "Creating PV for $DEVS"
pvcreate -y $DEVS
res=$?
if [ $res -ne 0 ]; then
    exit $res
fi

echo "Creating volume group k8s on $DEVS"
vgcreate k8s $DEVS

# only on master
# kubectl create -f deploy
