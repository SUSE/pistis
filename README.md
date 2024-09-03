# Pistis

Verifies the commits in a Git repository.

## Usage

In a given repository, a [CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners) file must contain patterns for protected paths along with users authorized to modify them. The CODEOWNERS syntax must use email addresses, not user names.
Along with the standardized `CODEOWNERS` file, a custom `CODEOWNERS_FINGERPRINTS` file must be present, containing the fingerprints for every email address contained in `CODEOWNERS`.
Both files are expected to be located at the repository root.

Commit hashes which should not be validated can be listed in a `TRUSTED_COMMITS` file in the repository root - one hash per line.

### Keyring

The verification requires a keyring containing the public keys of all `CODEOWNERS`, matching the fingerprints in `CODEOWNERS_FINGERPRINTS`.

#### Using a file

A keyring file can be passed using `--keyring` - it must be an armored export. Example to generate a suitable file:

```
gpg recv-key 9cf35828cec50de0294e04a1c645433b1e5e7a65
gpg recv-key 3DCEE0B7D6023F7B515FEF69244AE3A48488AFE5
gpg --export --armor -o /tmp/ring
```

Ideally, this should be a clean key ring. If run under a user account already having other keys imported, consider the `--keyring` argument with `gpg`.

#### Using GitLab

Alternatively to specifying an existing keyring file, a keyring can be built from PGP keys associated with users on a GitLab server.
For this, specify `--gitlab` instead of `--keyring`, and either maintain a `CODEOWNERS_USERNAMES` file containing mappings from the `CODEOWNERS` email addresses to GitLab usernames in the repository root, or provide a GitLab API token authorized to lookup users with their email address as `GITLAB_TOKEN` in the environment.

### Example `CODEOWNERS`

```
*.sls georg.pfuetzenreuter@suse.com
README.md cat@example.com
```

### Example `CODEOWNERS_FINGERPRINTS`

```
georg.pfuetzenreuter@suse.com 9cf35828cec50de0294e04a1c645433b1e5e7a65
cat@example.com 3DCEE0B7D6023F7B515FEF69244AE3A48488AFE5
```

### Example `CODEOWNERS_USERNAMES`

```
georg.pfuetzenreuter@suse.com crameleon
cat@example.com kitty
```

### Example `TRUSTED_COMMITS`

```
5ea8af38993138b5164451434b453ad9fd3993bd
38687faa23aeaf43e52cf513d761522fde4115e6
```

### Example application call

```
pistis --repository ~/Work/git/salt-crameleon/ --keyring /tmp/ring
```

or

```
pistis --repository ~/Work/git/salt-crameleon/ --gitlab https://gitlab.example.com
```

## TODO

- Change noisy messages to Debug()
- Verify signatures
- Consolidate `CODEOWNERS_FINGERPRINTS` and `CODEOWNERS_USERNAMES` into a single YAML file?
- Tests for main logic
- Cache/store last trusted commit, start next run from it
- Test mode running only new commits in branch
