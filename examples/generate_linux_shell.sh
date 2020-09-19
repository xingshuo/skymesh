#根据.bat推导生成对应.sh
for name in `ls *.bat`; do cp $name ${name%.bat}.sh; sed -i 's/\\/\//g' ${name%.bat}.sh; done