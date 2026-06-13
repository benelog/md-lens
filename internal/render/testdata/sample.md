# mdl Sample Document

A rich terminal markdown viewer. This paragraph mixes **bold**, *italic*,
`inline code`, ~~strikethrough~~, and a [project link](https://example.com).

## Lists

- First bullet
- Second bullet
  - Nested bullet
  - Another nested item
- Third bullet

1. First step
2. Second step
3. Third step

### Task list

- [x] Implement the parser
- [x] Render headings as images
- [ ] Add sixel support

## Blockquote

> The terminal is a canvas.
> Even large fonts can be drawn.

## Code

```java
// Greet the world
public class Hello {
    public static void main(String[] args) {
        String name = "mdl";
        System.out.println("Hello, " + name + "!");
    }
}
```

```python
def greet(name: str) -> str:
    # f-strings work too
    return f"Hello, {name}!"

print(greet("mdl"))
```

```json
{
  "name": "mdl",
  "version": "0.1.0",
  "images": true
}
```

```bash
# build and run
./gradlew shadowJar
java -jar build/libs/mdl.jar sample.md
```

```rust
fn main() {
    let name = "mdl";
    println!("Hello, {}!", name);
}
```

## Table

| Feature        | Status | Notes              |
|:---------------|:------:|-------------------:|
| Headings       |   ✓    |       font images  |
| Syntax colors  |   ✓    |        15 languages |
| Images         |   ✓    |   kitty/iterm/half |

## Image

![the mdl sample image](sample.png)

---

That's all — thanks for trying **mdl**!
