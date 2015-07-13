num=$(cat .posts | wc -l)

if [ -z $post ]; then
	post=$num
elif [ $post -gt $num ] || [ $post -lt 0 ]; then
	echo "No such post"
	exit
fi

if [ $post -lt $num ]; then
	next=$(( $post + 1 ))
fi

if [ $post -gt 1 ]; then
	prev=$(( $post - 1 ))
fi

file=`head -n$post .posts | tail -n1`

name=$(basename $file | sed "s/\.md//")
content=$(markdown $file)

echo "<div>$content</div>"

if [ ! -z $prev ]; then
	echo "<a href=\"index.sh?post=$prev\">Prev</a>"
fi
if [ ! -z $next ]; then
	echo "<a href=\"index.sh?post=$next\">Next</a>"
fi

echo "</body></html>"

