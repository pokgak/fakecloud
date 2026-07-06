# Lesson 1: your first fakecloud resource.
#
# Open http://localhost:8000 in a browser first, then:
#   terraform apply        -> a VM card appears on the dashboard
#   (edit instance_type)   -> apply again, watch the yellow ~ change in the feed
#   terraform destroy      -> the card disappears
#
# Then cause some drift: delete the VM with the dashboard's
# "delete out-of-band" button and run `terraform plan`.

terraform {
  required_providers {
    fakecloud = {
      source = "pokgak/fakecloud"
    }
  }
}

provider "fakecloud" {
  endpoint = "http://localhost:8000"
}

resource "fakecloud_vm" "web" {
  name          = "web-1"
  instance_type = "t2.micro"
}

output "vm_id" {
  value = fakecloud_vm.web.id
}
