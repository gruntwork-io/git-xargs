#!/usr/bin/env bash 

YEAR=$(date +"%Y")
FULLNAME="Gruntwork, LLC"

function create_license {
cat << EOF > LICENSE.txt
MIT License

Copyright (c) 2016 to $YEAR, $FULLNAME

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
}

# Copyrights should be declared as "$CREATION_YEAR to $CURRENT_YEAR"
# Therefore, this sed command will look to update the date immediately following the word "to"
function update_license_copyright_year {
	echo "Updating license copyright year to $(date +%Y)..."
	sed -i "s|to \([1-9][0-9][0-9][0-9]\)|to $(date +%Y)|" LICENSE.txt
	if [ $? -eq 0 ]; then 
		echo "Success!"
	else
		echo "Error!"
	fi	
}

# if the repo does not contain a LICENSE.txt file, then create one with the correct year
if [ ! -f "LICENSE.txt" ]; then 
	echo "Could not find LICENSE.txt at root of repo, so adding one..."
	create_license
else 
	update_license_copyright_year
fi 


