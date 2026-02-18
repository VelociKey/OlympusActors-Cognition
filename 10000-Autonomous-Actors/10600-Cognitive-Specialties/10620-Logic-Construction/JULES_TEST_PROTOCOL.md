# Protocol: Coder-to-Jules Test Generation

This protocol defines how the **Logic-Construction (10620)** agent tasks the **Jules Assistant** with background test generation.

## 1. Trigger Criteria
The CoderAgent should trigger a Jules campaign when:
- A new Go module is initialized.
- A ConnectRPC service definition is updated.
- Complexity in a `main.go` file exceeds established legibility thresholds.

## 2. Command Pattern
The CoderAgent executes the following Go-native command to task Jules:

```bash
jules /generate:tests --target="<Module_Path>" --type="unit,integration" --output="auto"
```

## 3. Verification Loop
1. **Tasking**: CoderAgent identifies the target and invokes Jules.
2. **Background Execution**: Jules generates `*_test.go` files asynchronously.
3. **Audit**: The Architectural-Synthesis (10610) agent verifies the new test coverage during the next fleet audit cycle.

*Status: ACTIVE*
