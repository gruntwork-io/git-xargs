#!/usr/bin/env ruby
# This script updates a repo to use Terraform 0.14. We are implementing the steps outlined in
# https://www.notion.so/gruntwork/Terraform-0-14-Upgrade-87f06063b1bd46789a6e0089168674d6

require 'json'
require 'open3'

DEFAULT_TERRAFORM_VERSION = "0.14.8"

REQUIRED_PROVIDERS_REGEX = /^\s*required_providers\s*{\s*$/

def get_root_folder
  `git rev-parse --show-toplevel`.strip
end

# Updates the Terraform version in CircleCI config file
def update_circleci_build_to_tf14 root_folder
  puts "Updating Terraform's version..."
  current_image = `yq eval '.defaults.docker[0].image' #{root_folder}/.circleci/config.yml`

  if current_image.strip != "null"
    `yq eval '.defaults.docker[0].image="087285199408.dkr.ecr.us-east-1.amazonaws.com/circle-ci-test-image-base:go1.16-go111module"' -i #{root_folder}/.circleci/config.yml`
  elsif File.readlines("#{root_folder}/.circleci/config.yml").grep(/TERRAFORM_VERSION:/).size > 0
    circle_ci_search_and_replace root_folder, /TERRAFORM_VERSION: .+/, "TERRAFORM_VERSION: #{DEFAULT_TERRAFORM_VERSION}"
    circle_ci_search_and_replace root_folder, /^(.*)GOLANG_VERSION: .+(\n.*GO111MODULE: auto.*)?/, '\1GOLANG_VERSION: 1.16\n\1GO111MODULE: auto'
    circle_ci_search_and_replace root_folder, '\n', "\n"
    circle_ci_search_and_replace root_folder, /^(.*)MODULE_CI_VERSION: .+/, '\1MODULE_CI_VERSION: v0.31.0'
  else
    raise "Did not find either a Docker image nor TERRAFORM_VERSION in CircleCI config"
  end
end

# Updates Go's version in every folder that has a go.mod
def update_terratest_in_go_mod root_folder
  paths_with_golang = Dir.glob "#{root_folder}/**/go.mod"

  paths_with_golang.each do |path|
    if File.readlines(path).grep(/terratest/).size > 0
      puts "Updating Terratest's version at #{path}..."

      dir = File.dirname(path)
      `cd #{dir} && go get -u github.com/gruntwork-io/terratest@v0.31.3`
      `cd #{dir} && go mod tidy`
    else
      puts "Found #{path} but it does not include Terratest as a dependency. Skipping."
    end
  end
end

# Updates the Terragrunt version in CircleCI config file
def update_circleci_build_to_tg27 root_folder
  terragrunt_version = `yq eval '.env.environment.TERRAGRUNT_VERSION' #{root_folder}/.circleci/config.yml`

  if (terragrunt_version.strip != "null") && (terragrunt_version.strip != "NONE")
    puts "Updating Terragrunt's version..."
    `yq eval -P '.env.environment.TERRAGRUNT_VERSION="v0.27.1"' -i #{root_folder}/.circleci/config.yml`
  else
    puts "Did not find a Terragrunt version (or it is set to NONE) in the CircleCi config. Skipping."
  end
end

# Removes `!!merge` tag left over from yq in the circleci/config.yml file
def remove_unused_yq_tags root_folder
  circle_ci_search_and_replace root_folder, "!!merge ", ""
end

# Updates the CircleCI config.yml file, replacing a given string to a new one.
def circle_ci_search_and_replace root_folder, old_str, new_str
  path_to_circle_config_yml = "#{root_folder}/.circleci/config.yml"
  circle_ci_config = File.read(path_to_circle_config_yml)
  sanitised_circle_ci_config = circle_ci_config.gsub(old_str, new_str)

  if sanitised_circle_ci_config != circle_ci_config
    File.open(path_to_circle_config_yml, "w") do |file|
     file.write(sanitised_circle_ci_config)
    end

    true
  else
    false
  end
end

# Add .terraform.lock.hcl to .gitignore, as our repos only have example code, and we want the latest providers on them,
# so there's no need to lock the versions.
def ignore_terraform_lock_file root_folder
  gitignore_path = "#{root_folder}/.gitignore"
  new_gitignore_value = %{
# Ignore Terraform lock files, as we want to test the Terraform code in these repos with the latest provider
# versions.
.terraform.lock.hcl
}

  if File.readlines(gitignore_path).grep(/.terraform.lock.hcl/).size == 0
    puts "Adding .terraform.lock.hcl to .gitignore"
    File.open(gitignore_path, 'a') do |file|
      file.write(new_gitignore_value)
    end
  else
    puts ".terraform.lock.hcl is already in .gitignore. Skipping."
  end
end

# Update the required_version constraint in the Terraform code. Also, our code has several flavors of comment blocks
# that talk about what version of Terraform the code works with or is tested with, so we update or clean up those
# comment blocks here too.
def update_tf_version_in_code root_folder
  regex_to_update = [
      {
          # Remove out of date comment blocks we forgot about during the 0.13 upgrade. These look something like:
          #
          # -----------------------------------------------------------------------------------------------
          # REQUIRE A SPECIFIC TERRAFORM VERSION OR HIGHER
          # This module uses HCL2 syntax, which means it is not compatible with any versions below 0.12.
          # -----------------------------------------------------------------------------------------------
          :regex => /#\s*-+\s*\n# REQUIRE A SPECIFIC TERRAFORM VERSION OR HIGHER\s*\n# This module uses HCL2 syntax, which means it is not compatible with any versions below 0.12.*\n#\s*-+\s*\n/,
          :replacement => ''
      },
      {
          # Remove out of date comment blocks we forgot about during the 0.13 upgrade. These look something like:
          #
          # -----------------------------------------------------------------------------------------------
          # REQUIRE A SPECIFIC TERRAFORM VERSION OR HIGHER
          # This module has been updated with 0.12 syntax, which means it is no longer compatible with any versions below 0.12.
          # -----------------------------------------------------------------------------------------------
          :regex => /#\s*-+\s*\n# REQUIRE A SPECIFIC TERRAFORM VERSION OR HIGHER\s*\n# This module has been updated with 0.12 syntax, which means it is no longer compatible with any versions below 0.12.*\n#\s*-+\s*\n/,
          :replacement => ''
      },
      {
          # Remove out of date comment blocks we forgot about during the 0.13 upgrade.
          :regex => /# This module has been updated with 0.12 syntax, which means it is no longer compatible with any versions below 0.12.*\n/,
          :replacement => ''
      },
      {
          # Update the first sentence in a comment block we added during the 0.13 upgrade to now talk about 0.14. These
          # look something like:
          #
          # This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting
          # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
          # forwards compatible with 0.13.x code.
          :regex => /This module is now only being tested with Terraform 0.13.x/,
          :replacement => 'This module is now only being tested with Terraform 0.14.x'
      },
      {
          # Update the last sentence in a comment block we added during the 0.13 upgrade to now talk about 0.14. These
          # look something like:
          #
          # This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting
          # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
          # forwards compatible with 0.13.x code.
          :regex => /forwards compatible with 0.13.x code/,
          :replacement => 'forwards compatible with 0.14.x code'
      }
  ]

  paths_with_terraform = Dir.glob "#{root_folder}/**/*.tf"
  paths_with_terraform.each do |path|
    contents = File.read(path)
    updated_contents = regex_to_update.reduce(contents) do |current_contents, to_update|
      current_contents.gsub(to_update[:regex], to_update[:replacement])
    end

    updated_contents = update_required_tf_version updated_contents, path
    updated_contents = ensure_comment_block_present updated_contents, path

    if contents != updated_contents
      puts "Updating Terraform version in comments in #{path}"
      File.write(path, updated_contents)
    else
      puts "Did not find Terraform version in any comments to update in #{path}"
    end
  end
end

# Ensure that the required_version in the given contents (which are assumed to be a Terraform file at the given path)
# is set to the value we are currently using (>= 0.12.26).
def update_required_tf_version contents, path
  expected_version_constraint = ">= 0.12.26"
  version_constraint_regex = /required_version\s*=\s*"(.+?)"/

  required_version = contents.match(version_constraint_regex)
  if required_version
    version_constraint = required_version.captures.first
    if version_constraint != expected_version_constraint
      puts "Updating version constraint in #{path} from #{version_constraint} to #{expected_version_constraint}"
      return contents.gsub(version_constraint_regex, "required_version = \"#{expected_version_constraint}\"")
    end
  end

  contents
end

# Ensure that we have a consistent comment block above the required_version constraint in the given contents  (which
# are assumed to be a Terraform file at the given path)
def ensure_comment_block_present contents, path
  comment_block = %{
  # This module is now only being tested with Terraform 0.14.x. However, to make upgrading easier, we are setting
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
  # forwards compatible with 0.14.x code.
}

  version_constraint_prev_line_regex = /^(.*)\n(\s*required_version\s*=\s*".+?")/

  required_version_and_prev_line = contents.match(version_constraint_prev_line_regex)
  if required_version_and_prev_line
    prev_line, required_version = required_version_and_prev_line.captures
    if prev_line.strip != "# forwards compatible with 0.14.x code."
      puts "Adding missing comment block above required_version constraint in #{path}"
      return contents.gsub(version_constraint_prev_line_regex, "#{prev_line}#{comment_block}#{required_version}")
    end
  end
  
  contents
end


# In TF 0.14, a provider { ... } block with a version = "<CONSTRAINT>" is deprecated and needs to be replaced with a
# required_providers block inside of a terraform { ... } block. This function uses a simple regex to look for these
# version constraints and exit with an error if it finds one so that the human operator can go and fix them.
def update_provider_constraints root_folder
  puts "Switching to Terraform #{DEFAULT_TERRAFORM_VERSION}"
  `tfenv install #{DEFAULT_TERRAFORM_VERSION}`
  `tfenv use #{DEFAULT_TERRAFORM_VERSION}`

  paths_with_terraform = Dir.glob "#{root_folder}/**/*.tf"
  folders_with_terraform = paths_with_terraform.map { |path| File.dirname(path) }.uniq

  folders_with_terraform.each do |folder|
    # We make the (hopefully not too inaccurate) assumption that all provider and terraform blocks in our modules are
    # defined in a main.tf file. If there's no main.tf file, we just skip it.
    main_tf_path = File.join(folder, "main.tf")
    if !File.exist?(main_tf_path)
      next
    end

    main_tf_contents = IO.read(main_tf_path)

    blocks = hcledit(["block", "list"], main_tf_contents).split("\n")

    has_terraform_block = blocks.any? { |block| block == "terraform" }
    if !has_terraform_block
      main_tf_contents = add_terraform_block(main_tf_contents, main_tf_path)
    end

    providers = blocks.select { |block| block.start_with?("provider.") }.map{ |block| block.gsub(/^provider./, "") }
    providers.each do |provider|
      main_tf_contents = update_provider_constraint(provider, main_tf_contents, main_tf_path)
    end

    IO.write(main_tf_path, main_tf_contents)
  end
end

def update_provider_constraint(provider, main_tf_contents, main_tf_path)
  version_constraint = hcledit(["attribute", "get", "provider.#{provider}.version"], main_tf_contents).strip
  if version_constraint.length == 0
    return main_tf_contents
  end

  puts "Removing version constraint for provider #{provider} in #{main_tf_path}"
  main_tf_contents = hcledit(["attribute", "rm", "provider.#{provider}.version"], main_tf_contents)

  has_required_providers_constraint = hcledit(["attribute", "get", "terraform.required_providers.#{provider}"], main_tf_contents)
  if has_required_providers_constraint.length > 0
    puts "There is already a required_providers version constraint for provider #{provider} in #{main_tf_path}, so will not add another one."
  else
    terraform_block = hcledit(["block", "get", "terraform"], main_tf_contents)
    if !REQUIRED_PROVIDERS_REGEX.match(terraform_block)
      puts "Adding required_providers block to #{main_tf_path}"
      main_tf_contents = hcledit(["block", "append", "terraform", "required_providers", "--newline"], main_tf_contents)
    end

    puts "Adding version constraint for provider #{provider} to required_providers block in #{main_tf_path}"
    version_constraint_block = """{
  source  = \"hashicorp/#{provider}\"
  version = #{version_constraint}
}
"""
    main_tf_contents = hcledit(["attribute", "append", "terraform.required_providers.#{provider}", version_constraint_block], main_tf_contents)
  end

  main_tf_contents
end

# Add a new terraform { ... } block. If the file starts with a comment block, we try to put the terraform { ... } block
# after that comment block; otherwise, we put it at the top of the file.
def add_terraform_block(main_tf_contents, main_tf_path)
  puts "Adding terraform { ... } block to #{main_tf_path}"

  lines = main_tf_contents.split("\n")
  line_no = 0
  while line_no < lines.length && lines[line_no].start_with?("#")
    line_no = line_no + 1
  end

  terraform_block = """
terraform {
  # This module is now only being tested with Terraform 0.14.x. However, to make upgrading easier, we are setting
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
  # forwards compatible with 0.14.x code.
  required_version = \">= 0.12.26\"
}
"""

  lines.insert(line_no, terraform_block)
  lines.join("\n")
end

# Run hcledit with the given args and the given stdin. Return stdout.
def hcledit(args, stdin)
  out, err, status = Open3.capture3("hcledit", *args, stdin_data: stdin)
  if status != 0
    raise "hcledit exited with exit code #{status}: #{err}"
  end
  out
end

def update_readme_badge root_folder
  # In root/README.md or root/README.adoc:
  # replace string https://img.shields.io/badge/tf-%3E%3D0.12.0-blue.svg
  # with    string https://img.shields.io/badge/tf-%3E%3D0.14.0-blue.svg

  tf_12_badge = 'https://img.shields.io/badge/tf-%3E%3D0.12.0-blue.svg'
  tf_14_badge = 'https://img.shields.io/badge/tf-%3E%3D0.14.0-blue.svg'
  readme_adoc_path = root_folder + '/README.adoc'
  readme_md_path = root_folder + '/README.md'

  for readme_path in [readme_adoc_path, readme_md_path]
    if File.exist?(readme_path)
      puts "Found " + readme_path
      readme_contents = File.read(readme_path)
      if readme_contents.include? tf_12_badge
        puts "Found TF 12 badge, replacing with TF 14 badge."
        readme_contents = readme_contents.gsub(tf_12_badge, tf_14_badge)
        IO.write(readme_path, readme_contents)
      else
        puts "Did not find TF 12 badge. Skipping."
      end
    else
      puts "Did not find " + readme_path + '. Skipping.'
    end
  end
end

def terraform_format root_folder
  `cd #{root_folder} && terraform fmt -recursive .`
end

root_folder = get_root_folder
update_circleci_build_to_tf14 root_folder
update_terratest_in_go_mod root_folder
update_circleci_build_to_tg27 root_folder
remove_unused_yq_tags root_folder
ignore_terraform_lock_file root_folder
update_tf_version_in_code root_folder
update_provider_constraints root_folder
update_readme_badge root_folder
terraform_format root_folder

