# MCP Integration

ccflow workflows can integrate with MCP (Model Context Protocol) servers for enhanced functionality. This guide covers recommended setups.

## Overview

MCP servers provide Claude Code with access to external tools and data sources. ccflow doesn't automatically configure MCP, but records your preferences and provides setup guidance.

## Supported Integrations

### Version Control

#### GitHub

**Server**: GitHub MCP Server
**Repository**: https://github.com/modelcontextprotocol/servers/tree/main/src/github

**Setup**:
1. Install the server:
   ```bash
   npm install -g @modelcontextprotocol/server-github
   ```

2. Configure in Claude Code settings:
   ```json
   {
     "mcpServers": {
       "github": {
         "command": "npx",
         "args": ["-y", "@modelcontextprotocol/server-github"],
         "env": {
           "GITHUB_PERSONAL_ACCESS_TOKEN": "<your-token>"
         }
       }
     }
   }
   ```

**Capabilities**:
- Create and manage pull requests
- Review PR comments
- Search issues
- Access repository information

#### GitLab

**Server**: GitLab MCP Server (community)

**Setup**: Check MCP server directory for GitLab implementations.

### Issue Tracking

#### Linear

**Server**: Linear MCP Server
**Repository**: Check MCP server directory

**Setup**:
```json
{
  "mcpServers": {
    "linear": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-linear"],
      "env": {
        "LINEAR_API_KEY": "<your-api-key>"
      }
    }
  }
}
```

**Capabilities**:
- Create and update issues
- Search issues
- Manage project workflows

#### Jira

**Server**: Jira MCP Server (community)

**Setup**: Check MCP server directory for Jira implementations.

## Configuration in ccflow

When running `ccflow run`, you'll be asked about MCP preferences:

```
? Version control system (for MCP guidance): GitHub
? Issue tracker (for MCP guidance): Linear
```

These choices are recorded in `workflow.yaml`:

```yaml
mcp:
  vcs: github
  tracker: linear
  deploy: none
```

ccflow uses these to:
1. Provide relevant setup instructions
2. Include MCP-aware examples in command templates
3. Suggest integrations in documentation

## Best Practices

### Security

1. **Never commit tokens**: Use environment variables or secure secret storage
2. **Minimal permissions**: Request only necessary scopes for tokens
3. **Rotate regularly**: Update API keys periodically

### Performance

1. **Cache appropriately**: MCP servers may cache data; understand refresh behavior
2. **Batch operations**: Group related operations when possible
3. **Handle errors**: MCP failures shouldn't break your workflow

### Workflow Integration

The ccflow commands can work with MCP servers:

- `/idea` can create issues in your tracker
- `/design` can link to PRs for review
- `/release` can update issue statuses

To enable these integrations, modify the command templates to include MCP server calls.

## Resources

- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [MCP Server Directory](https://github.com/modelcontextprotocol/servers)
- [Claude Code MCP Documentation](https://docs.anthropic.com/en/docs/claude-code/mcp)
