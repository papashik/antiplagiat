import sys
from PyGram import Profile, Node

ind = 0

if len(sys.argv) != 3:
	print("give two arguments")
	exit(1)

with open(sys.argv[1], 'r') as fp:
	str1 = fp.read()

with open(sys.argv[2], 'r') as fp:
	str2 = fp.read()

def parsing(node, string):
	global ind
	while not(string[ind]==')' and string[ind-1]!=' '):
		if (string[ind+1]=='(' and string[ind+2]!=' '):
			if (string[ind+2:string.find(' ', ind+1, len(string))]=='eos'):
				ind = string.find(')', ind+1, len(string))+1
			else:
				node.addkid(Node(string[ind+2:string.find(' ', ind+1, len(string))]))
				ind = string.find(' ', ind+1, len(string))
				parsing(node.children[-1], string)
		elif ((string[ind+1]=='(' and string[ind+2]==' ') or (string[ind]==' ' and string[ind+1]==')')):
			ind += 2
		else:
			if (string.find(' ', ind+1, len(string))==-1 or string.find(')', ind+1, len(string))<string.find(' ', ind+1, len(string))):
				if (string[ind+1:string.find(')', ind+1, len(string))]!='<EOF>'):
					node.addkid(Node(string[ind+1:string.find(')', ind+1, len(string))]))
				ind = string.find(')', ind+1, len(string))
			else:
				node.addkid(Node(string[ind+1:string.find(' ', ind+1, len(string))]))
				ind = string.find(' ', ind+1, len(string))
	ind += 1

def buildTree(string):
	global ind
	A = Node(string[1:string.find(' ')])
	ind = string.find(' ')
	parsing(A, string)
	return A

Tree1 = buildTree(str1)
ind = 0
Tree2 = buildTree(str2)

profile_a = Profile(Tree1)
profile_b = Profile(Tree2)

dist = profile_a.edit_distance(profile_b)
print(dist, end="")

exit(0)
