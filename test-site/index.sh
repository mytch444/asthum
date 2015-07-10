if [ -z $1 ]; then
	num=`cat .posts | wc -l`
else
	num=`echo $1 | cut -d'=' -f2`
fi

last=`cat .posts | wc -l`
if [ $num -lt $last ]; then
	next=$(( $num + 1 ))
fi

if [ $num -gt 1 ]; then
	prev=$(( $num - 1 ))
fi

file=`head -n$num .posts | tail -n1`

name=$(basename $file | sed "s/\.md//")
content=$(markdown $file)

echo "<div>$content</div>"

if [ ! -z $prev ]; then
	echo "<a href=\"index.sh?num=$prev\">Prev</a>"
fi
if [ ! -z $next ]; then
	echo "<a href=\"index.sh?num=$next\">Next</a>"
fi

echo "</body></html>"

