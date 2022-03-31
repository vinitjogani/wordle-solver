package main

import (
  "strings"
  "fmt"
  "os"
  "bufio"
  "math"
  "sort"
)

func read_lines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}


func filter_word(
  word string, 
  missing []IndexedChar, 
  exact []IndexedChar, 
  kinda []IndexedChar) bool {
  for _, char := range missing {
    if strings.Contains(word, char.char) {
       return false 
    }
  }

  for _, char := range exact {
    if string(word[char.i]) != char.char {
       return false 
    }
  }

  for _, char := range kinda {
    if string(word[char.i]) == char.char || !strings.Contains(word, char.char) {
       return false 
    }
  }

  return true
}


func compute_feedback(word string, actual string) string {
  out := ""
  for i, char := range word {
    if rune(actual[i]) == char { 
      out += "+" 
    } else if strings.Contains(actual, string(char)) { 
      out += "~" 
    } else { 
      out += "-" 
    }
  }
  return out
}

func compute_score(word string, dictionary []string, msg_out chan<- Result) {
  if word == "" { return }
  N := float64(len(dictionary))
  
  feedback_counts := make(map[string]int)
  for _, actual := range dictionary {
    feedback := compute_feedback(word, actual)
    val, ok := feedback_counts[feedback]
    if ok {
      feedback_counts[feedback] = val + 1
    } else {
      feedback_counts[feedback] = 1
    }
  }

  sum := 0.0
  for _, counts := range feedback_counts {
    p := float64(counts) / N
    sum += (1 - p) * p
  }

  msg_out <- Result{ Word: word, Eliminations: sum  }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}


func filter_words(
  all_words []string, 
  word string,
  feedback string) []string {
  out := []string{}

  missing := []IndexedChar{}
  exact := []IndexedChar{}
  kinda := []IndexedChar{}
  for i, c := range word {
    char := IndexedChar{i: i, char: string(c)}
    if feedback[i] == '-' {
      missing = append(missing, char)
    } else if feedback[i] == '+' {
      exact = append(exact, char)
    } else {
      kinda = append(kinda, char)
    }
  }

  for _, word := range all_words {
    if filter_word(word, missing, exact, kinda) {
      out = append(out, word)
    }
  }
  return out
}

func iter(all_words []string, filtered []string) {
  messages := make(chan Result)
  parallelism := 100
  N := float64(len(all_words))
  per_routine := int(math.Ceil(N / float64(parallelism)))

  total := 0
  for i := 0; i < parallelism; i++ {
    start := i*per_routine
    end :=  min(len(all_words), (i+1)*per_routine)
    slice := all_words[start:end]
    go func(words []string) {
      for _, word := range words {
        compute_score(word, filtered, messages)
      }
    }(slice)
    total += len(slice)
  }

  out := []Result{}
  for output := range messages {
    out = append(out, output)
    if len(out) % 500 == 0 {
      fmt.Println(
        len(out), "/", len(all_words), 
        output.Word, output.Eliminations)
    }
    if len(out) == total {
      close(messages)
      break
    }
  }

  sort.Slice(out, func(i, j int) bool {
		return out[i].Eliminations > out[j].Eliminations
	})

  fmt.Println(out[:10])
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func main() {
  all_words, _ := read_lines("words.txt")
  filtered := all_words

  for {
    fmt.Print("guess: ")
    var word string
    fmt.Scanln(&word)
    fmt.Print("feedback: ")
    var feedback string
    fmt.Scanln(&feedback)

    filtered = filter_words(filtered, word, feedback)
    if len(filtered) == 1 {
      fmt.Println(filtered[0])
      break
    }
    
    fmt.Println(len(filtered))
    fmt.Println(filtered[:min(len(filtered), 5)])
    iter(all_words, filtered)
  }
}

