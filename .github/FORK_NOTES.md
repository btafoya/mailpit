# Fork Maintenance Notes

## About This Fork
This is an independent fork of [btafoya/mailsandbox](https://github.com/btafoya/mailsandbox) with additional enterprise features.

## Enhanced Features
- **Postmark API Emulation**: Full compatibility with Postmark's API for testing
- **MCP Server Support**: Integration with Claude Code and other MCP clients
- **Docker Enhancements**: Improved Docker support with MCP integration
- **Extended Documentation**: Comprehensive guides for all new features

## Upstream Relationship
- Original: https://github.com/btafoya/mailsandbox
- This fork: https://github.com/btafoya/mailsandbox
- Status: Independent mainment with selective upstream merging

## Sync Strategy
```bash
# Fetch upstream changes (monthly)
git fetch upstream

# Review changes
git log upstream/main --oneline

# Cherry-pick specific fixes
git cherry-pick <commit>

# Or merge carefully
git merge upstream/main --no-ff
```

## Version Strategy
- Follow upstream version for base features
- Add suffix for fork features (e.g., v1.27.7-postmark.1)

## Contribution Guidelines
- PRs welcome for fork-specific features
- Consider upstream contribution for general fixes
- Maintain compatibility where possible

## Last Upstream Sync
- Date: Check git log
- Version: v1.27.7
- Next Review: Monthly
