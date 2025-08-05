class KubectlKruise < Formula
  desc "kubectl plugin for OpenKruise"
  homepage "https://openkruise.io/"
  version "1.2.2"
  license "Apache-2.0"

  on_macos do
    on_intel do
      url "https://github.com/openkruise/kruise-tools/releases/download/v1.2.2/kubectl-kruise-darwin-amd64-v1.2.2.tar.gz"
      sha256 "bfb978a55a8048a4b41b02ce9afec2f1e74718998e3a549dfd0a89f95dcec98b"
    end

    on_arm do
      url "https://github.com/openkruise/kruise-tools/releases/download/v1.2.2/kubectl-kruise-darwin-arm64-v1.2.2.tar.gz"
      sha256 "9be1ff3d3a86dc28db3f4513462a3f43536f8beb60b536c3a678b9c3387afed0"
    end
  end

  def install
    bin.install "kubectl-kruise"
    
    # Generate shell completions
    generate_completions_from_executable(bin/"kubectl-kruise", "completion")
  end

  test do
    system "#{bin}/kubectl-kruise", "version"
    
    # Test that the binary is executable and shows help
    output = shell_output("#{bin}/kubectl-kruise --help")
    assert_match "kubectl-kruise controls the OpenKruise CRs", output
    
    # Test completion generation
    output = shell_output("#{bin}/kubectl-kruise completion bash")
    assert_match "# bash completion", output
  end
end
