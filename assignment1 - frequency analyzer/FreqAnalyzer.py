from operator import itemgetter #for sorting entries by key/value
import time #for timing
"""Tree dictionaries will be used for keeping track of the 
    occurences of unigrams, bigrams, and trigrams respectively """
s_time = time.time();
dict_unigram = {}
dict_bigram = {}
dict_trigram = {}

book = open ("ebook2.txt", "r") #open the file in read mode
 #read file line by line and count occurences of unigrams, bigrams, and trigrams
line = book.readline().lower()
while line != '':
    line = line.strip()         #remove all leading and trailing white space
    for i in range(len(line)):  #process the whole line character by character
        if line[i] in dict_unigram:        #check if the character read is in the dictionary
                dict_unigram[line[i]] += 1 #increment its counter
        else:
            dict_unigram[line[i]] = 1 #add the character to the dictionary

        if i < len(line)-1 and line[i:i+2] in dict_bigram:
            dict_bigram[line[i:i+2]] += 1
        else:
            dict_bigram[line[i:i+2]] = 1
    
        if i < len(line)-2 and line[i:i+3] in dict_trigram:
            dict_trigram[line[i:i+3]] += 1
        else:
            dict_trigram[line[i:i+3]] = 1
    line = book.readline().lower() #the count is case insensitive
book.close()

def printTop20(Map, title):                         #for printing the results in the required format
    arr = sorted(Map.items(), key=itemgetter(0))    #sort the map alphanumerically (because the map has no specific order)
    arr.sort(key=itemgetter(1),reverse=True)        #sort the resulting array with respect to the frequency (the sort is stable)
    print (title)
    for i,tup in enumerate(arr):                    #print the 20 most frequent elements
        print ("\"" + tup[0] + "\"\t" + str(tup[1]))
        if i == 19:
            break
    print()

printTop20 (dict_unigram, "Unigrams:")
printTop20 (dict_bigram, "Bigrams:")
printTop20 (dict_trigram, "Trigrams:")


print("%.3f s" % (time.time() - s_time)) # print time elapsed in seconds