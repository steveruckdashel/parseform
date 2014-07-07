package main

import (
  "io"
  "bufio"
  )

// itemType identifies the type of lex items.
type itemType int

const (
    itemError itemType = iota // error occurred;
                              // value is text of error
    itemDot                   // the cursor, spelled '.'
    itemEOF
    itemElse       // else keyword
    itemEnd        // end keyword
    itemField      // identifier, starting with '.'
    itemIdentifier // identifier
    itemIf         // if keyword
    itemLeftMeta   // left meta-string
    itemNumber     // number
    itemPipe       // pipe symbol
    itemRange      // range keyword
    itemRawString  // raw quoted string (includes quotes)
    itemRightMeta  // right meta-string
    itemString     // quoted string (includes quotes)
    itemText       // plain text

    itemStartTag  // <
    itemEndTag    // >
    itemCloseTag  // '/'
    itemTagName    // tag name
    itemAtributeName // attribute name
    itemAtributeValue // attribute value

    itemEq
)

type item struct {
  typ itemType  // Type, such as itemNumber.
  val string    // Value, such as "23.2".
}

func (i item) String() string {
    switch i.typ {
    case itemEOF:
        return "EOF"
    case itemError:
        return i.val
    }
    if len(i.val) > 10 {
        return fmt.Sprintf("%.10q...", i.val)
    }
    return fmt.Sprintf("%q", i.val)
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// run lexes the input by executing state functions
// until the state is nil.
func run() {
    for state := lexText; state != nil; {
        state = state(l)
    }
    close(l.items) // No more tokens will be delivered.
}

// lexer holds the state of the scanner.
type lexer struct {
    stream *bufio.Reader

    name  string    // used only for error reports.
    input string    // the string being scanned.
    items chan item // channel of scanned items.
}

func lex(name string, input io.Reader) (*lexer, chan item) {
    l := &lexer{
        name:  name,
        stream: bufio.NewReader(input),
        items: make(chan item),
    }
    go l.run()  // Concurrently run state machine.
    return l, l.items
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
    l.items <- item{t, l.input}
    l.input = ""
}

func lexText(l *lexer) stateFn {
    for {
        if next, n, err := l.stream.ReadRune(); err==io.EOF || n==0 {
          l.emit(itemText)
          break
        } if next=='<' {
          l.emit(itemText)
          l.input = next
          return lexTag
        } else {
          l.input = append(l.input, next)
        }
    }
    l.emit(itemEOF)
    return nil       // Stop the run loop.
}

func lexTag(l *lexer) stateFn {
  l.emit(itemStartTag)
  if next, _, _ := l.stream.Peak(1); next=='/' {
    l.emit(itemCloseTag)
  }
  for {
      if next, n, err := l.stream.ReadRune(); err==io.EOF || n==0 {
        l.emit(itemEOF)
        break
      } else if !unicode.IsLetter(next) {
        if e := l.stream.UnreadRune(); e!=nil {
          log.Printf("%v", e)
        }
        l.emit(itemTagName)
      } else {
        l.input = append(l.input, next)
      }
  }
  return lexAttributes       // Stop the run loop.
}

func lexAttributes(l *lexer) stateFn {
  next, _, err := l.stream.ReadRune()
  if err==io.EOF {
    l.emit(itemEOF)
    return nil
  } else if err!=nil {
    log.Printf("%v",e)
    return nil
  }

  switch next {
    case '>':
      l.emit(itemEndTag)
      return lexText
    case '/':
      l.emit(itemCloseTag)
      return lexAttributes
    case '=':
      l.emit(itemEq)
      return lexAttributes
    default:
      if !unicode.IsLetter(next) {
        return lexAttributes
      }
  }

  for ;unicode.IsLetter(next); {
    l.input = append(l.input, next)
    next, _, err := l.stream.ReadRune()
  }
  l.emit(itemAttributeName)

  return lexAttributes
}

func main() {

}
