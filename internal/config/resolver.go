package config

import "fmt"

func (cfg *EcsConfig) ActiveContext() (*Context, error) {
	if cfg.CurrentContext == "" {
		if len(cfg.Contexts) == 0 {
			return nil, fmt.Errorf("no contexts configured; run: ecsctl config set-context <name> --cluster <name> --region <region>")
		}
		c := cfg.Contexts[0].Context
		return &c, nil
	}
	for _, entry := range cfg.Contexts {
		if entry.Name == cfg.CurrentContext {
			c := entry.Context
			return &c, nil
		}
	}
	return nil, fmt.Errorf("context %q not found; run: ecsctl config get-contexts", cfg.CurrentContext)
}

func (cfg *EcsConfig) SetContext(name string, ctx Context) {
	for i, entry := range cfg.Contexts {
		if entry.Name == name {
			cfg.Contexts[i].Context = ctx
			return
		}
	}
	cfg.Contexts = append(cfg.Contexts, ContextEntry{Name: name, Context: ctx})
}

func (cfg *EcsConfig) DeleteContext(name string) bool {
	for i, entry := range cfg.Contexts {
		if entry.Name == name {
			cfg.Contexts = append(cfg.Contexts[:i], cfg.Contexts[i+1:]...)
			return true
		}
	}
	return false
}
