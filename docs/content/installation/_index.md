---
title: "Installation"
description: "Installing the Autopilot CLI"
weight: 4
---

To install the Autopilot CLI, simply run the following:

```bash
curl -sL https://run.solo.io/autopilot/install | sh
export PATH=$HOME/.autopilot/bin:$PATH
```

Verify that `ap` installed correctly:
```bash
ap --version
```

```
autopilot version 0.0.1
```

Great! You're all set to start building operators. If you're just getting started with Autopilot, check out the [Getting Started Tutorial]({{< versioned_link_path fromRoot="/tutorial_code/getting_started_1">}})
