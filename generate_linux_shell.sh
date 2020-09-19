#根据.bat推导生成对应.sh
cp -f build.bat build.sh
sed -i 's/\\/\//g' build.sh
sed -i 's/\@echo/echo/g' build.sh
sed -i 's/del /rm /g' build.sh
for name in `ls examples/*.bat`; do cp $name ${name%.bat}.sh; sed -i 's/\\/\//g' ${name%.bat}.sh; done