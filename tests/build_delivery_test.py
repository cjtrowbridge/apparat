"""Exercise first delivery, interruption, no-op and push rejection against local Git.
The fixture supplies a deterministic compiler and download channel without devices.
"""
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from scripts import build_delivery as delivery
from scripts.build_orchestrator import BuildPlan


class DeliveryTest(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)/'checkout'
        self.remote = self.root.parent/'remote.git'
        self.root.mkdir()
        self.git('init','-b','main')
        self.git('config','user.name','Fixture')
        self.git('config','user.email','fixture@invalid')
        self.git('init','--bare','-b','main',self.remote)
        self.git('remote','add','origin',self.remote)
        self.git('commit','--allow-empty','-m','initial')
        self.git('push','origin','main')
        self.output = 'releases/linux/amd64/apparat/latest'
        self.plan = BuildPlan('linux/amd64/apparat','linux','amd64','apparat',self.root/self.output,True)
        self.storage = self.root/'.tmp'
        self.storage.mkdir()
        self.builds = 0

    def git(self,*args):
        return subprocess.check_output(['git','-C',self.root,*args],stderr=subprocess.PIPE).decode().strip()

    def qualify(self,*args):
        self.builds += 1
        path = self.root/self.output
        path.parent.mkdir(parents=True,exist_ok=True)
        path.write_bytes(b'qualified')
        row = dict(name=self.plan.name,output=self.output,state='passed',sha256=delivery.evidence.digest(path))
        (self.root/delivery.REPORT).write_text(json.dumps(dict(accepted=True,inputs='inputs',targets=[row])))
        return 0

    def verify(self,path,digest):
        data = subprocess.check_output(['git','--git-dir',self.remote,'show','main:'+path])
        self.assertEqual(hashlib.sha256(data).hexdigest(),digest)

    def run_pipeline(self,qualify=None,verify=None):
        publish = delivery.publish
        with patch.object(delivery.evidence,'locked_run',side_effect=qualify or self.qualify), patch.object(
                delivery.evidence,'fingerprint',return_value='inputs'), patch.object(
                delivery,'publish',side_effect=lambda root,rows: publish(root,rows,verify or self.verify)):
            return delivery.deliver(self.root,[self.plan],None,self.storage)

    def test_resume_after_push_and_repeat(self):
        def interrupted(*args):
            raise OSError('network interrupted')
        with self.assertRaisesRegex(OSError,'network'):
            self.run_pipeline(verify=interrupted)
        head = self.git('rev-parse','HEAD')
        self.assertEqual(self.run_pipeline(),0)
        self.assertEqual(self.run_pipeline(),0)
        self.assertEqual(self.git('rev-parse','HEAD'),head)
        self.assertEqual(self.builds,1)

    def test_failed_target_then_rejected_push(self):
        head = self.git('rev-parse','HEAD')
        self.assertEqual(self.run_pipeline(qualify=lambda *args: 1),1)
        self.assertEqual(self.git('rev-parse','HEAD'),head)
        hook = self.remote/'hooks/pre-receive'
        hook.write_text('#!/bin/sh\nexit 1\n')
        hook.chmod(0o755)
        with self.assertRaises(subprocess.CalledProcessError):
            self.run_pipeline()
        self.assertEqual(self.git('ls-remote','origin','refs/heads/main').split()[0],head)
        hook.unlink()
        self.assertEqual(self.run_pipeline(),0)
        self.assertEqual(self.builds,1)
