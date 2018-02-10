#include <stdio.h> // for fopen, fclose, fgets, puts, printf
#include <string.h> // for strcpy, strcmp, strncpy, strlen
#include <stdbool.h> // for bool type
#include <stdlib.h> // for malloc, free
#include <ctype.h> // for tolower
#include <time.h> // for timing

#define BUFLEN 5000
typedef struct Entry
{
	char *key;
	long val;
} Entry;
typedef struct AVLNode
{
	struct Entry *entry;
	unsigned int h;
    struct AVLNode *left;
    struct AVLNode *right;
} AVLNode;
typedef struct AVLTree
{
	AVLNode *root;
	int size;
} AVLTree;
/*
avl tree (used as dictionary) the inorder traversal gives the entries sorted alphanumerically
*/
//adds a new key or updates its count if it already exists in the tree
void add(AVLTree *tree, const char *newkey, const int n); 
// recursive add
AVLNode* addR(AVLNode *node, const char *newkey, bool *newOne, const int n);
AVLNode* updateHeight(AVLNode* node);   // used for keeping the balance of the AVL tree
AVLNode* rightRotate(AVLNode* node);    //  ~
AVLNode* leftRotate(AVLNode* node);     // ~
int checkBalance(AVLNode* node);        // ~
// deallocates the space used by the tree
void clear(AVLTree *tree);
AVLNode* clearR(AVLNode *node); // clear helper

/*
 Functions for printing the 20 most frequent entries
*/
//copies the AVL tree into an array (sorted by key)
int avl2arr(struct Entry** arr, AVLNode *node, int index);
//sorts the entries with respect to val. The sort is stable so the alphanumeric order
//is preserved for entries with the same count.
void printTop20(AVLTree avl, char *title);

/*
 Merge-sort functions
*/
void SplitMerge(Entry **data, Entry **help, long iBegin, long iEnd);
void MergeArrays(Entry **data, Entry **help, long iBegin, long iMiddle, long iEnd);
void Copy(Entry **data, Entry **help, long iBegin, long iEnd);
void MergeSort(Entry **data, long size);
int Compare(Entry* e1, Entry* e2);

int main ()
{
    clock_t begin = clock(); // start timing
	FILE *book = fopen ("ebook.txt", "r");
    AVLTree unigrams, bigrams, trigrams; // "dictionaries"

    unigrams.root = NULL;
    bigrams.root = NULL;
    trigrams.root = NULL;

	char line[BUFLEN];
	char uni[2], bi[3], tri[4];
	uni[1] = bi[2] = tri[3] = '\0';
    //read file line by line and count unigrams, bigrams and trigrams
	while(fgets(line, BUFLEN, book) != NULL)
	{
        int start = 0;
        char *temp = line;
        while(temp[0] == ' ') temp++;// trim leading spaces
        unsigned int len = strlen(temp);
        if(temp[len-1] == '\n') temp[--len] = '\0';// remove new line character
            
        for(int i = 0; temp[i]; i++)
          temp[i] = tolower(temp[i]); // remove case sensitivity
        if(len > 2)
        {
            for(int i = 0; i < len-2; i++)
            {
        		strncpy(uni,&temp[i],1); //get substring (copies character at i)
                add (&unigrams, uni, 2);
        		
                strncpy(bi,&temp[i],2); //get substring (copies substring [i,i+1])
                add (&bigrams, bi, 3);
        		
                strncpy(tri,&temp[i],3); //get substring (copies substring [i,i+2])
                add (&trigrams, tri, 4);
            }
        }
        if(len >= 2)
        {
            strncpy(uni,&temp[len-2],1);
            add (&unigrams, uni, 2);
            strncpy(bi,&temp[len-2],2);
            add (&bigrams, bi, 3);

            strncpy(uni,&temp[len-1],1);
            add (&unigrams, uni, 2);
        }
        else if(len >= 1)
        {
            strncpy(uni,&temp[len-1],1);
            add (&unigrams, uni, 2);
        }
	}
    fclose(book);
    
    printTop20(unigrams, "Unigrams:");
    clear(&unigrams);

    printTop20(bigrams, "Bigrams:");
    clear(&bigrams);
    
    printTop20(trigrams, "Trigrams:");
    clear(&trigrams);
	
    clock_t end = clock();
    float time_spent = (float)(end - begin) / CLOCKS_PER_SEC;
    printf("%.3f s",time_spent); // print time spent in seconds
    return 0;
}

void add(AVLTree *tree, const char *newkey, const int n)
{
    if(tree->root == NULL) // the tree is empty create the root node
    {
        tree->root = (AVLNode*) malloc(sizeof(AVLNode));
        tree->root->entry = (struct Entry*) malloc (sizeof(struct Entry));
        tree->root->entry->key = (char*) malloc (n*sizeof(char));
        strcpy(tree->root->entry->key, newkey);
        tree->root->entry->val = 1;
        tree->size = 1;
        tree->root->left = tree->root->right = NULL;
    }
    else
    {
        bool newOne = false; //checks if a new entry was added or a value was incremented
        tree->root = addR(tree->root, newkey, &newOne,n);
        if(newOne) tree->size ++;
    }
}

AVLNode* addR(AVLNode *node, const char *newkey, bool *newOne, const int n)
{
    if(node == NULL) // newkey does not exist in the tree, add it
    {
        node = malloc(sizeof(AVLNode));
        node->entry = (struct Entry*) malloc (sizeof(struct Entry));
        node->entry->key = (char*) malloc (n*sizeof(char));
        strcpy(node->entry->key, newkey);
        node->entry->val = 1;
        node->left = node->right = NULL;
        *newOne = true;
        return node;
    }
    if(strcmp(newkey, node->entry->key) == 0) // newkey is found, increment its counter
    {
        node->entry->val ++;
        return node;
    }
    else if(strcmp(newkey, node->entry->key) < 0) // new key is smaller than the current key, go left
    {
        node->left = addR(node->left, newkey, newOne, n);
        if(checkBalance(node) < -1)//node->left->h - node->right->h > 1)
        {
            if(checkBalance(node->left) > 1)//node->left->right->h - node->left->left->h > 1) //add some conditions for NULLpointers
                node->left = leftRotate(node->left); //don't forget to update the height of modified nodes
            node = rightRotate(node);
        }
        node = updateHeight(node);
        return node;
    }
    else // new key is larger than the current key, go right
    {
        node->right = addR(node->right, newkey, newOne, n);
        if(checkBalance(node) > 1)//node->right->h - node->left->h > 1)
        {
            if(checkBalance(node->right) < -1)
                node->right = rightRotate(node->right);
            node = leftRotate(node);
        }
        node = updateHeight(node);
        return node;
    }
}

#define max(x, y) ((x>y)?x:y)
AVLNode* updateHeight(AVLNode* node) // updates the height of a node after an addition
{
    if (node == NULL || (node->left == NULL && node->right == NULL))
        return node;
    if(node->left == NULL)
        node->h = 1 + node->right->h;
    else if(node->right == NULL)
        node->h = 1 + node->left->h;
    else
        node->h = 1+max(node->left->h, node->right->h);
    return node;
}
AVLNode* rightRotate(AVLNode* node)
{
    AVLNode *temp = node->left;
    node->left = temp->right;
    node = updateHeight(node);
    temp->right = node;
    node = temp;
    node = updateHeight(node);
    return node;
}
AVLNode* leftRotate(AVLNode* node)
{
    AVLNode *temp = node->right;
    node->right = temp->left;
    node = updateHeight(node);
    temp->left = node;
    node = temp;
    node = updateHeight(node);
    return node;
}
int checkBalance(AVLNode* node) // checks if the avl tree is balanced
{
    if(node == NULL || (node->left == NULL && node->right == NULL))
        return 0;
    if(node->left == NULL)
        return ((node->right->h <= 1)?0:2);
    if(node->right == NULL)
        return ((node->left->h <= 1)?0:-2);
    return (node->right->h - node->left->h);
}
void clear(AVLTree *tree) //clears all values in the tree
{
    if(tree != NULL)
    {
        tree->root = clearR(tree->root);
    }
}
AVLNode* clearR(AVLNode *node) // recursive method to clear all nodes starting from the leaves
{
    if(node == NULL)
        return node;
    node->left = clearR(node->left);
    node->right = clearR(node->right);
    free(node->entry);
    free(node->left);
    free(node->right);
    return node;
}
int avl2arr(struct Entry** arr, AVLNode *node, int index) // converts the avl tree into a tree sorted by key
{
    if(node == NULL)
        return index;
    index = avl2arr(arr, node->left, index);
    arr[index] = node->entry;
    index++;
    index = avl2arr(arr, node->right, index);
    return index;
}
void printTop20(AVLTree avl, char *title) // prints the 20 most frequent entries
{
    struct Entry **arr;
    arr = malloc (avl.size*sizeof(struct Entry*));
    avl2arr(arr, avl.root, 0);
    MergeSort(arr, avl.size);
    puts(title);
    int n = (avl.size >= 20)?20:avl.size;
    for(int i = 0; i < n; i++)
    {
        printf("\"%s\"\t%lu\n",arr[i]->key,arr[i]->val);
    }
    free(arr);
}

/**
sorts an array by sorting its halves and merging them
*/
void MergeSort(Entry **data, long size)
{
    Entry **help = malloc(size*sizeof(Entry*));
    *help = malloc(size*sizeof(Entry));
    for (int i = 0; i < size; i++)
        help[i] = 0L;
    SplitMerge(data, help, 0, size);
}
/**
divides the array in half each time to simplify the problem
*/
void SplitMerge(Entry **data, Entry **help, long iBegin, long iEnd){
    if (iEnd - iBegin < 2)                       // if run size == 1
        return;                                 //   consider it sorted
    // recursively split runs into two halves until run size == 1,
    // then merge them and return back up the call chain
    long iMiddle = (iEnd + iBegin) / 2;              // iMiddle = mid point
    SplitMerge(data, help, iBegin, iMiddle);  // split / merge left  half
    SplitMerge(data, help, iMiddle, iEnd);  // split / merge right half
    MergeArrays(data, help, iBegin, iMiddle, iEnd);  // merge the two half runs
    Copy(data, help, iBegin, iEnd);              // copy the merged runs back to A
}

/**
*/
void MergeArrays(Entry **data, Entry **help, long iBegin, long iMiddle, long iEnd){
    long i0 = iBegin, i1 = iMiddle;

    // While there are elements in the left or right runs
    for (long j = iBegin; j < iEnd; j++) {
        // If left run head exists and is <= existing right run head.
        if (i0 < iMiddle && (i1 >= iEnd || Compare(data[i0], data[i1]) <= 0)){
            help[j] = data[i0];
            i0 = i0 + 1;
        }
        else{
            help[j] = data[i1];
            i1 = i1 + 1;
        }
    }
}

/**
copies the sorted help array into the inputted array
*/
void Copy(Entry **data, Entry **help, long iBegin, long iEnd){
    for (long k = iBegin; k < iEnd; k++)
        data[k] = help[k];
}
int Compare(Entry* e1, Entry* e2)
{
    return e2->val - e1->val;
}