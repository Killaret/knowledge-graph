.PHONY: clean-lunix clean-lunix-sh clean-lunix-dry

# Clean and compress lunix image (convenience targets)
clean-lunix:
	@echo "Running clean:lunix (PowerShell) via npm"
	npm run clean:lunix

clean-lunix-sh:
	@echo "Running clean:lunix:sh (bash) via npm"
	npm run clean:lunix:sh

clean-lunix-dry:
	@echo "Dry run (PowerShell)"
	npm run clean:lunix:dry

# If needed, user can call make clean-lunix-sh-dry manually using the script flags
clean-lunix-sh-dry:
	@echo "Dry run (bash)"
	npm run clean:lunix:sh:dry
