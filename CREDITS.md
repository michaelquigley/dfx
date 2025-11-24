# Credits and Acknowledgments

dfx is built on top of several excellent open source projects and freely available fonts. We are grateful to the creators and maintainers of these projects.

## Core Dependencies

### Dear ImGui
**Author:** Omar Cornut (ocornut)
**Repository:** https://github.com/ocornut/imgui
**License:** MIT License

Dear ImGui is the foundational immediate-mode GUI library that powers dfx. It provides bloat-free graphical user interfaces for C++ with minimal dependencies.

> Dear ImGui is designed to enable fast iterations and to empower programmers to create content creation tools and visualization / debug tools (as opposed to UI for the average end-user). It favors simplicity and productivity toward this goal, and lacks certain features normally found in more high-level libraries.

### cimgui-go
**Author:** AllenDang
**Repository:** https://github.com/AllenDang/cimgui-go
**License:** MIT License

cimgui-go provides Go bindings for Dear ImGui through cimgui, making it possible to use Dear ImGui from Go applications. This library handles the complex C/Go interop and provides a Go-friendly API.

## Fonts

### Gidole Regular
**Designer:** Andreas Larsen
**Website:** https://gidole.github.io/
**License:** Open Font License (OFL)

Gidole is a free, open-source, modern DIN font. It is used as the primary UI font in dfx for its clean, readable appearance.

### JetBrains Mono
**Author:** JetBrains
**Repository:** https://github.com/JetBrains/JetBrainsMono
**License:** OFL-1.1 (SIL Open Font License 1.1)

JetBrains Mono is a typeface specifically designed for developers. It is used in dfx for displaying monospace text, code, and logs. Features include increased height for better readability and distinctive letterforms to reduce eye strain.

### Material Icons
**Author:** Google
**Repository:** https://github.com/google/material-design-icons
**License:** Apache License 2.0

Material Icons are the official icon set from Google's Material Design. dfx uses the Material Icons Regular font for consistent, recognizable iconography throughout the UI.

## Additional Dependencies

dfx also depends on several Go libraries:

- **github.com/michaelquigley/df** - a dynamic application framework for golang applications
- **github.com/pkg/errors** - Enhanced error handling with stack traces
- **golang.design/x/clipboard** - Cross-platform clipboard access
- **github.com/sqweek/dialog** - Native file dialogs

## License

dfx itself is licensed under the Apache License 2.0. See the LICENSE file for details.

## Contributing

If you use dfx in your project, we encourage you to maintain similar attribution for these foundational projects and assets.
