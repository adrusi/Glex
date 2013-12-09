# Glex
Glex implements a lexigraphic analyzer (scanner) for go, serving the same
purpose as Flex. It makes heavy use of go's reflection capabilities and focuses on ease of use and elegance. It strives to have full support for utf8 input.

## Development Status
Glex is in very early stages of development, and is not fit for use by even those who like to code on the bleeding edge. I doubt that it is capable of lexing grammars with medium to high complexity, and I'm sure that it's riddled with bugs. That said, the project is very new and will hopefully mature quickly.

### Priorities
1. **Tests** Test coverage is very sparse at the moment, the situation should improve before proceeding much further.
2. **Documentation** Even in this early, feature-lacking state, the documentation is even more lacking. I want to get an early start on a proper documentation effort.
3. **Features** Glex is currently nowhere near as powerful as tools like Flex. It needs to be able to lex any grammar for it to be useful.

## Usage
For now, see example/main.go for example usage.
