"""
    Allows for the computation of the PQ-Gram edit distance of two trees. To calculate the distance,
    a Profile object must first be created for each tree, then the edit_distance function can be called.

    For more information on the PQ-Gram algorithm, please see the README.
"""

import copy

#####################################################
################### file "tree" #####################
#####################################################


class Node(object):
    """
        A generic representation of a tree node. Includes a string label and a list of a children.
    """

    def __init__(self, label):
        """
            Creates a node with the given label. The label must be a string for use with the PQ-Gram
            algorithm.
        """
        self.label = label
        self.children = list()

    def addkid(self, node, before=False):
        """
            Adds a child node. When the before flag is true, the child node will be inserted at the
            beginning of the list of children, otherwise the child node is appended.
        """
        if before:  self.children.insert(0, node)
        else:   self.children.append(node)
        return self

##### Helper Methods #####

def split_tree(root, delimiter=""):
    """
        Traverses a tree and explodes it based on the given delimiter. Each node is split into a null
        node with each substring as a separate child. For example, if a node had the label "A:B:C" and
        was split using the delimiter ":" then the resulting node would have "*" as a parent with the
        children "A", "B", and "C". By default, this explodes each character in the label as a separate
        child node. Relies on split_node.
    """
    if(delimiter == ''):
        sub_labels = [x for x in root.label]
    else:
        sub_labels = root.label.rsplit(delimiter)
    if len(sub_labels) > 1: # need to create a new root
        new_root = Node("*", 0)
        for label in sub_labels:
            new_root.children.append(Node(label, 0))
        heir = new_root.children[0]
    else: # root wasn't split, use it as the new root
        new_root = Node(root.label, 0)
        heir = new_root
    for child in root.children:
        heir.children.extend(split_node(child, delimiter))
    return new_root

def split_node(node, delimiter):
    """
        Splits a single node into children nodes based on the delimiter specified.
    """
    if(delimiter == ''):
        sub_labels = [x for x in node.label]
    else:
        sub_labels = node.label.rsplit(delimiter)
    sub_nodes = list()
    for label in sub_labels:
        sub_nodes.append(Node(label, 0))
    if len(sub_nodes) > 0:
        for child in node.children:
            sub_nodes[0].children.extend(split_node(child, delimiter))
    return sub_nodes

#####################################################
#####################################################
#####################################################

class Profile(object):
    """
        Represents a PQ-Gram Profile, which is a list of PQ-Grams. Each PQ-Gram is represented by a
        ShiftRegister. This class relies on both the ShiftRegister and tree.Node classes.
    """

    def __init__(self, root, p=2, q=3):
        """
            Builds the PQ-Gram Profile of the given tree, using the p and q parameters specified.
            The p and q parameters do not need to be specified, however, different values will have
            an effect on the distribution of the calculated edit distance. In general, smaller values
            of p and q are better, though a value of (1, 1) is not recommended, and anything lower is
            invalid.
        """
        super(Profile, self).__init__()
        ancestors = ShiftRegister(p)
        self.list = list()

        self.profile(root, p, q, ancestors)
        self.sort()

    def profile(self, root, p, q, ancestors):
        """
            Recursively builds the PQ-Gram profile of the given subtree. This method should not be called
            directly and is called from __init__.
        """
        ancestors.shift(root.label)
        siblings = ShiftRegister(q)

        if(len(root.children) == 0):
            self.append(ancestors.concatenate(siblings))
        else:
            for child in root.children:
                siblings.shift(child.label)
                self.append(ancestors.concatenate(siblings))
                self.profile(child, p, q, copy.deepcopy(ancestors))
            for i in range(q-1):
                siblings.shift("*")
                self.append(ancestors.concatenate(siblings))

    def edit_distance(self, other):
        """
            Computes the edit distance between two PQ-Gram Profiles. This value should always
            be between 0.0 and 1.0. This calculation is reliant on the intersection method.
        """
        union = len(self) + len(other)
        return 1.0 - 2.0*(self.intersection(other)/union)

    def intersection(self, other):
        """
            Computes the set intersection of two PQ-Gram Profiles and returns the number of
            elements in the intersection.
        """
        intersect = 0.0
        i = j = 0
        while i < len(self) and j < len(other):
            intersect += self.gram_edit_distance(self[i], other[j])
            if self[i] == other[j]:
                i += 1
                j += 1
            elif self[i] < other[j]:
                i += 1
            else:
                j += 1
        return intersect

    def gram_edit_distance(self, gram1, gram2):
        """
            Computes the edit distance between two different PQ-Grams. If the two PQ-Grams are the same
            then the distance is 1.0, otherwise the distance is 0.0. Changing this will break the
            metrics of the algorithm.
        """
        distance = 0.0
        if gram1 == gram2:
            distance = 1.0
        return distance

    def sort(self):
        """
            Sorts the PQ-Grams by the concatenation of their labels. This step is automatically performed
            when a PQ-Gram Profile is created to ensure the intersection algorithm functions properly and
            efficiently.
        """
        self.list.sort(key=lambda x: ''.join(x))

    def append(self, value):
        self.list.append(value)

    def __len__(self):
        return len(self.list)

    def __repr__(self):
        return str(self.list)

    def __str__(self):
        return str(self.list)

    def __getitem__(self, key):
        return self.list[key]

    def __iter__(self):
        for x in self.list: yield x

class ShiftRegister(object):
    """
        Represents a register which acts as a fixed size queue. There are only two valid
        operations on a ShiftRegister: shift and concatenate. Shifting results in a new
        value being pushed onto the end of the list and the value at the beginning list being
        removed. Note that you cannot recover this value, nor do you need to for generating
        PQ-Gram Profiles.
    """

    def __init__(self, size):
        """
            Creates an internal list of the specified size and fills it with the default value
            of "*". Once a ShiftRegister is created you cannot change the size without
            concatenating another ShiftRegister.
        """
        self.register = list()
        for i in range(size):
            self.register.append("*")

    def concatenate(self, reg):
        """
            Concatenates two ShiftRegisters and returns the resulting ShiftRegister.
        """
        temp = list(self.register)
        temp.extend(reg.register)
        return temp

    def shift(self, el):
        """
            Shift is the primary operation on a ShiftRegister. The new item given is pushed onto
            the end of the ShiftRegister, the first value is removed, and all items in between shift
            to accomodate the new value.
        """
        self.register.pop(0)
        self.register.append(el)

"""
    The following methods are provided for visualization of the PQ-Gram Profile structure. They
    are NOT intended for other use, and play no role in using the PQ-Gram algorithm.
"""

def build_extended_tree(root, p=1, q=1):
    """
        This method will take a normal tree structure and the given values for p and q, returning
        a new tree which represents the so-called PQ-Extended-Tree.

        To do this, the following algorithm is used:
            1) Add p-1 null ancestors to the root
            2) Traverse tree, add q-1 null children before the first and
               after the last child of every non-leaf node
            3) For each leaf node add q null children
    """
    original_root = root # store for later

    # Step 1
    for i in range(p-1):
        node = tree.Node(label="*")
        node.addkid(root)
        root = node

    # Steps 2 and 3
    list_of_children = original_root.children
    if(len(list_of_children) == 0):
        q_append_leaf(original_root, q)
    else:
        q_append_non_leaf(original_root, q)
    while(len(list_of_children) > 0):
        temp_list = list()
        for child in list_of_children:
            if(child.label != "*"):
                if(len(child.children) == 0):
                    q_append_leaf(child, q)
                else:
                    q_append_non_leaf(child, q)
                    temp_list.extend(child.children)
        list_of_children = temp_list
    return root

##### Extended Tree Functions #####

def q_append_non_leaf(node, q):
    """
        This method will append null node children to the given node. (Step 2)

        When adding null nodes to a non-leaf node, the null nodes should exist on both side of
        the real children. This is why the first of each pair of children added sets the flag
        'before=True', ensuring that on the left and right (or start and end) of the list of
        children a node is added.
    """
    for i in range(q-1):
        node.addkid(tree.Node("*"), before=True)
        node.addkid(tree.Node("*"))

def q_append_leaf(node, q):
    """
        This method will append q null node children to the given node. (Step 3)
    """
    for i in range(q):  node.addkid(tree.Node("*"))
