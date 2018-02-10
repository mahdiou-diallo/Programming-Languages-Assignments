//for file reading
import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
//for dictionary sorted by key
import java.util.TreeMap;
import java.util.Map.Entry;
//for sorting the final array by value
import java.util.Arrays;
import java.util.Comparator;

public class FreqAnalyzer
{
	public static void main(String[] args) throws IOException
	{
		long s_time = System.currentTimeMillis();
		//Dictionaries for storing counts of unigrams, bigrams, and trigrams
		TreeMap <Character,Integer> unigram_map = new TreeMap <Character, Integer>();
		TreeMap <String,Integer> bigram_map = new TreeMap <String, Integer>();
		TreeMap <String,Integer> trigram_map = new TreeMap <String, Integer>();
		BufferedReader book = new BufferedReader(new FileReader("ebook2.txt"));
		
		String line = "";
		//process the file line by line and count
		while ((line = book.readLine()) != null)
		{
			line = line.toLowerCase().trim(); // remove leading and trailing white space
			Character uni;
			String bi = "", tri = "";
			for (int i = 0; i < line.length(); i++)
			{
				uni = line.charAt(i);
				add(uni,unigram_map);
				if(i < line.length()-1)
				{
					bi = line.substring(i,i+2);
					add(bi,bigram_map);

					if(i < line.length()-2)
					{
						tri = line.substring(i,i+3);
						add(tri,trigram_map);
					}
				}
			}
		}
		book.close();
		printTop20(unigram_map, "Unigrams:");
		printTop20(bigram_map, "Bigrams:");
		printTop20(trigram_map, "Trigrams:");

		System.out.println(((System.currentTimeMillis() - s_time)/1000.0) + " s"); //print duration
	}
	static void add(Character c, TreeMap <Character,Integer> map)
	{
		if(map.containsKey(c)) // check if the value exists in the map
		{
			int val = map.get (c); // get previous counter value
			map.put (c,val+1); //update the counter
		}
		else
			map.put (c,1); // add the value to the map
	}
	static void add(String str, TreeMap <String,Integer> map)
	{
		if(map.containsKey(str)) // check if the value exists in the map
		{
			int val = map.get (str);
			map.put (str,val+1); //update the counter
		}
		else
			map.put (str,1); // add the value to the map
	}
	public static void printTop20(TreeMap map, String title)
	{
		Entry[] arr = (Entry[])map.entrySet().toArray(new Entry[map.size()]); // create array from treemap
		Comparator<Entry> comparator = new Comparator<Entry>() // custom comparator for sorting
	    {
	        public int compare(Entry e1, Entry e2)
	        {
	            return ((Integer)e2.getValue()).compareTo((Integer)e1.getValue()) ; // e2 is used first for reversing the order
	        }
	    };
		Arrays.sort(arr, comparator);
	    System.out.println(title);
	    int n = Math.min(arr.length, 20);
	    for (int i = 0; i < n; i++) // print most frequent entries
	        System.out.println("\"" + arr[i].getKey() + "\"\t" + arr[i].getValue());
	    System.out.println();
	}
}