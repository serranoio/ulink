# Ulink

Tool for switching between relative paths & node_modules

## Usecase

- If you have a package that you want to link locally and still want to benefit from hot reloading, use this tool to replace all references to node_modules to the relative path of your file

# Example

```
ui/
├── src/
│   ├── components/
│   │   ├── Button/
│   │   │   ├── Button.tsx
│   └── utils/
│       ├── helpers.ts
```

```
library/
├── src/
│   ├── components/
│   │   ├── Square/
│   │   │   ├── Square.tsx
│   └── utils/
│       ├── helpers.ts
```

## Before

In both files, the import will be resolved from node_modules as simply `library`

```
import { Square } from library
```

## After

`ui/src/components/Button/Button.tsx` ->

```
import { Square } from '../../../../library/src/components/Square/Square.tsx
```

`ui/src/components` ->

```
import { Square } from '../../../../library/src/components/Square/Square.tsx
```
