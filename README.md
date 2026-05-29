<p align="center">
  <img src="orion_logo.png" alt="Orion logo" width="280">
</p>

# Orion → Go Transpiler

Transpilador da linguagem **Orion** para **Go**, escrito em Go.

## Como usar

```bash
# Construir o transpilador
go build -o orionc ./cmd/orion/main.go

# Transpilar um arquivo .ori
./orionc meu_programa.ori

# Transpilar e executar
./orionc -run meu_programa.ori

# Ver tokens e AST (debug)
./orionc -debug meu_programa.ori
```

---

## Sintaxe Orion

### Variáveis
```orion
string name = "hermes";
integer age = 45;
bool is_male = true;
float pi = 3.14;
```

### Condicionais (`if / or_if / or`)
```orion
if is_male {
    io::write("is male");
} or_if age > 30 {
    io::write("older");
} or {
    io::write("default");
}
```

### Funções com retorno
```orion
string helloName (string name) {
    ::("hello, $name !");   // retorno implícito com interpolação
}
string hello = helloName("hermes");
```

### Funções void
```orion
writeScreen () {
    io::write("hello, world");
}
writeScreen();
```

### Arrays
```orion
[string] names = ["hermes", "lucas"];

// Iteração
for index, value :: names {
    io::write(value);
}

// Métodos de array
names:push("gusta");
names:pop();
names:first();
names:last();
names:remove(0);
names:removeValue("hermes");
```

### Interpolação de strings
```orion
string greeting = "hello, $name !";
```

### I/O
```orion
io::write("hello, world");
io::write(variable);
```

---

## Estrutura do projeto

```
orion/
├── cmd/orion/main.go          # CLI
├── internal/
│   ├── lexer/
│   │   ├── token.go           # Definição de tokens
│   │   └── lexer.go           # Tokenizador
│   ├── parser/
│   │   └── parser.go          # Parser (tokens → AST)
│   ├── ast/
│   │   └── ast.go             # Nós da AST
│   └── codegen/
│       └── codegen.go         # Gerador de código Go
├── example.ori                # Exemplo de código Orion
└── go.mod
```

## Mapeamento de tipos

| Orion     | Go        |
|-----------|-----------|
| `string`  | `string`  |
| `integer` | `int`     |
| `bool`    | `bool`    |
| `float`   | `float64` |
| `[T]`     | `[]T`     |
